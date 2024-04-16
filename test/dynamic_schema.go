package main

import (
	"context"   // 引入context包，用于控制子程序的执行、取消、超时等
	"flag"      // 引入flag包，用于解析命令行参数
	"fmt"       // 引入fmt包，用于格式化输出
	"log"       // 引入log包，用于记录日志
	"math/rand" // 引入math/rand包，用于生成随机数
	"time"      // 引入time包，用于处理时间相关的功能

	"github.com/milvus-io/milvus-sdk-go/v2/client" // 引入Milvus SDK的客户端包
	"github.com/milvus-io/milvus-sdk-go/v2/entity" // 引入Milvus SDK的实体定义包
)

// 定义常量
const (
	milvusAddr     = `你的Milvus服务的地址:19530` // Milvus服务的地址
	nEntities, dim = 3000, 128             // 插入的实体数量和向量维度
	collectionName = "dynamic_example"     // 集合名称

	// 用于格式化输出的消息模板
	msgFmt                                             = "\n==== %s ====\n"
	idCol, typeCol, randomCol, sourceCol, embeddingCol = "ID", "type", "random", "source", "embeddings"
	topK                                               = 4 // 查询时返回的相似度最高的前K个结果
)

func main() {
	flag.Parse()                // 解析命令行参数
	ctx := context.Background() // 创建一个背景上下文

	fmt.Printf(msgFmt, "start connecting to Milvus") // 输出连接Milvus的开始消息
	// 创建Milvus客户端，如果失败则记录日志并退出
	c, err := client.NewGrpcClient(ctx, milvusAddr)
	if err != nil {
		log.Fatalf("failed to connect to milvus, err: %v", err)
	}
	defer c.Close() // 确保在函数返回前关闭客户端连接

	// 获取并输出Milvus服务的版本信息
	version, err := c.GetVersion(ctx)
	if err != nil {
		log.Fatal("failed to get version of Milvus server", err.Error())
	}
	fmt.Println("Milvus Version:", version)

	// 检查指定的集合是否存在，如果存在则删除
	has, err := c.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("failed to check collection exists, err: %v", err)
	}
	if has {
		c.DropCollection(ctx, collectionName)
	}

	// 创建集合
	fmt.Printf(msgFmt, "create collection `dynamic_example")
	schema := entity.NewSchema().
		WithName(collectionName).
		WithDescription("dynamic schema example collection").
		WithAutoID(false).
		WithDynamicFieldEnabled(true).
		WithField(entity.NewField().WithName(idCol).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName(embeddingCol).WithDataType(entity.FieldTypeFloatVector).WithDim(dim))

	if err := c.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil { // 使用默认分片数创建集合
		log.Fatalf("create collection failed, err: %v", err)
	}

	// 描述集合
	fmt.Printf(msgFmt, "describe collection `dynamic_example`")
	coll, err := c.DescribeCollection(ctx, collectionName)
	if err != nil {
		log.Fatal("failed to describe collection:", err.Error())
	}

	fmt.Printf("Collection %s\tDescription: %s\tDynamicEnabled: %t\n", coll.Schema.CollectionName, coll.Schema.CollectionName, coll.Schema.EnableDynamicField)
	for _, field := range coll.Schema.Fields {
		fmt.Printf("Field: %s\tDataType: %s\tIsDynamic: %t\n", field.Name, field.DataType.String(), field.IsDynamic)
	}

	// 插入数据
	fmt.Printf(msgFmt, "start inserting with extra columns")
	idList, randomList := make([]int64, 0, nEntities), make([]float64, 0, nEntities)
	typeList := make([]int32, 0, nEntities)
	embeddingList := make([][]float32, 0, nEntities)

	rand.Seed(time.Now().UnixNano())
	// 生成数据
	for i := 0; i < nEntities; i++ {
		idList = append(idList, int64(i))
		typeList = append(typeList, int32(i%3))
	}
	for i := 0; i < nEntities; i++ {
		randomList = append(randomList, rand.Float64())
	}
	for i := 0; i < nEntities; i++ {
		vec := make([]float32, 0, dim)
		for j := 0; j < dim; j++ {
			vec = append(vec, rand.Float32())
		}
		embeddingList = append(embeddingList, vec)
	}

	// 使用生成的数据进行插入操作
	idColData := entity.NewColumnInt64(idCol, idList)
	randomColData := entity.NewColumnDouble(randomCol, randomList)
	typeColData := entity.NewColumnInt32(typeCol, typeList)
	embeddingColData := entity.NewColumnFloatVector(embeddingCol, dim, embeddingList)
	if _, err := c.Insert(ctx, collectionName, "", idColData, randomColData, typeColData, embeddingColData); err != nil {
		log.Fatalf("failed to insert random data into `dynamic_example, err: %v", err)
	}

	// 刷新集合，确保数据写入磁盘
	if err := c.Flush(ctx, collectionName, false); err != nil {
		log.Fatalf("failed to flush data, err: %v", err)
	}
	fmt.Printf(msgFmt, "start inserting with rows")

	// 通过结构体插入数据
	type DynamicRow struct {
		entity.RowBase
		ID     int64     `milvus:"name:ID;primary_key"`     // 主键字段
		Vector []float32 `milvus:"name:embeddings;dim:128"` // 向量字段
		Source int32     `milvus:"name:source"`             // 源字段
		Value  float64   `milvus:"name:random"`             // 随机值字段
	}

	// 生成并插入结构体数据
	rows := make([]entity.Row, 0, nEntities)
	for i := 0; i < nEntities; i++ {
		vec := make([]float32, 0, dim)
		for j := 0; j < dim; j++ {
			vec = append(vec, rand.Float32())
		}

		rows = append(rows, &DynamicRow{
			ID:     int64(nEntities + i),
			Vector: vec,
			Source: 1,
			Value:  rand.Float64(),
		})
	}

	_, err = c.InsertByRows(ctx, collectionName, "", rows)
	if err != nil {
		log.Fatal("failed to insert by rows: ", err.Error())
	}

	// 使用map[string]interface{}进行数据插入
	fmt.Printf(msgFmt, "start to inserting by MapRow")
	m := make(map[string]interface{})
	m["ID"] = int64(nEntities)
	vec := make([]float32, 0, dim)
	for j := 0; j < dim; j++ {
		vec = append(vec, rand.Float32())
	}
	m["embeddings"] = vec
	m["source"] = int32(1)
	m["random"] = rand.Float64()

	_, err = c.InsertByRows(ctx, collectionName, "", []entity.Row{entity.MapRow(m)})
	if err != nil {
		log.Fatal("failed to insert by rows: ", err.Error())
	}

	// 再次刷新数据
	if err := c.Flush(ctx, collectionName, false); err != nil {
		log.Fatalf("failed to flush data, err: %v", err)
	}

	// 创建索引
	fmt.Printf(msgFmt, "start creating index IVF_FLAT")
	idx, err := entity.NewIndexIvfFlat(entity.L2, 128)
	if err != nil {
		log.Fatalf("failed to create ivf flat index, err: %v", err)
	}
	if err := c.CreateIndex(ctx, collectionName, embeddingCol, idx, false); err != nil {
		log.Fatalf("failed to create index, err: %v", err)
	}

	// 加载集合到内存
	fmt.Printf(msgFmt, "start loading collection")
	start := time.Now()
	err = c.LoadCollection(ctx, collectionName, false)
	if err != nil {
		log.Fatalf("failed to load collection, err: %v", err)
	}

	fmt.Printf("load collection done, time elapsed: %v\n", time.Since(start))
	fmt.Printf(msgFmt, "start searching based on vector similarity")

	// 执行基于向量相似度的搜索
	vec2search := []entity.Vector{
		entity.FloatVector(embeddingList[len(embeddingList)-2]),
		entity.FloatVector(embeddingList[len(embeddingList)-1]),
	}
	begin := time.Now()
	sp, _ := entity.NewIndexIvfFlatSearchParam(16)
	sRet, err := c.Search(ctx, collectionName, nil, "", []string{randomCol, typeCol}, vec2search,
		embeddingCol, entity.L2, topK, sp)
	end := time.Now()
	if err != nil {
		log.Fatalf("failed to search collection, err: %v", err)
	}

	fmt.Println("results:")
	for _, res := range sRet {
		printResult(&res, map[string]string{randomCol: "double", typeCol: "int"})
	}
	fmt.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)

	// 执行混合搜索
	fmt.Printf(msgFmt, "start hybrid searching with `random > 0.9`")
	begin = time.Now()
	sRet2, err := c.Search(ctx, collectionName, nil, "random > 0.9",
		[]string{randomCol, typeCol}, vec2search, embeddingCol, entity.L2, topK, sp)
	end = time.Now()
	if err != nil {
		log.Fatalf("failed to search collection, err: %v", err)
	}
	fmt.Println("results:")
	for _, res := range sRet2 {
		printResult(&res, map[string]string{randomCol: "double", typeCol: "int"})
	}
	fmt.Printf("\tsearch latency: %dms\n", end.Sub(begin)/time.Millisecond)

	// 执行查询操作
	expr := "ID in [0, 1, 2]"
	fmt.Printf(msgFmt, fmt.Sprintf("query with expr `%s`", expr))
	sRet3, err := c.Query(ctx, collectionName, nil, expr, []string{randomCol, typeCol})
	if err != nil {
		log.Fatalf("failed to query result, err: %v", err)
	}
	printResultSet(sRet3, map[string]string{idCol: "int", randomCol: "double", typeCol: "int"})

	// 执行复杂查询
	expr = "source in [1] and random > 0.1"
	fmt.Printf(msgFmt, fmt.Sprintf("query with expr `%s`", expr))
	sRet3, err = c.Query(ctx, collectionName, nil, expr, []string{randomCol, typeCol, sourceCol}, client.WithLimit(3))
	if err != nil {
		log.Fatalf("failed to query result, err: %v", err)
	}
	printResultSet(sRet3, map[string]string{idCol: "int", randomCol: "double", typeCol: "int", sourceCol: "int"})

	// 删除集合
	fmt.Printf(msgFmt, "drop collection `dynamic_example`")
	if err := c.DropCollection(ctx, collectionName); err != nil {
		log.Fatalf("failed to drop collection, err: %v", err)
	}
}

// printResultSet 函数用于打印查询结果集
func printResultSet(sRets client.ResultSet, outputInfo map[string]string) {
	for name, typ := range outputInfo {
		column := sRets.GetColumn(name)
		if column == nil {
			fmt.Printf("column %s not found in result set\n", name)
			continue
		}

		fmt.Printf("Result Column %s, count: %d\n", name, column.Len())
		switch typ {
		case "int":
			// 输出整型数据
			var data []int64
			for i := 0; i < column.Len(); i++ {
				line, err := column.GetAsInt64(i)
				if err != nil {
					fmt.Printf("failed to get column %s at index %d, err: %s\n", name, i, err.Error())
				}
				data = append(data, line)
			}
			fmt.Println("Data:", data)
		case "string":
			// 输出字符串数据
			var data []string
			for i := 0; i < column.Len(); i++ {
				line, err := column.GetAsString(i)
				if err != nil {
					fmt.Printf("failed to get column %s at index %d, err: %s\n", name, i, err.Error())
				}
				data = append(data, line)
			}
			fmt.Println("Data:", data)
		case "double":
			// 输出浮点数数据
			var data []float64
			for i := 0; i < column.Len(); i++ {
				line, err := column.GetAsDouble(i)
				if err != nil {
					fmt.Printf("failed to get column %s at index %d, err: %s\n", name, i, err.Error())
				}
				data = append(data, line)
			}
			fmt.Println("Data:", data)
		}
	}
}

// printResult 函数用于打印搜索结果
func printResult(sRet *client.SearchResult, outputInfo map[string]string) {
	printResultSet(sRet.Fields, outputInfo)
}
