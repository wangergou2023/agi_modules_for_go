package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// 定义扑克牌的花色符号和点数
var suits = []string{"♥", "♦", "♣", "♠"}
var ranks = []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}

// 定义大小王
var jokers = []string{"🃏大", "🃏小"}

// 定义花色和点数的优先级，用于排序
var suitOrder = map[string]int{"♥": 1, "♦": 2, "♣": 3, "♠": 4}
var rankOrder = map[string]int{
	"2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6, "8": 7, "9": 8, "10": 9, "J": 10, "Q": 11, "K": 12, "A": 13,
}

// Card 结构体表示一张扑克牌
type Card struct {
	Suit string
	Rank string
}

// 返回扑克牌的字符串表示（使用符号）
func (c Card) String() string {
	if c.Suit == "" {
		return c.Rank // 处理大小王的情况
	}
	return fmt.Sprintf("%s%s", c.Suit, c.Rank)
}

// 构建一副牌（根据参数决定是否包含大小王）
func createDeck(includeJokers bool) []Card {
	var deck []Card
	// 添加普通牌
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	// 添加大小王
	if includeJokers {
		for _, joker := range jokers {
			deck = append(deck, Card{Suit: "", Rank: joker})
		}
	}
	return deck
}

// 洗牌函数
func shuffle(deck []Card) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range deck {
		j := rng.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

// 排序函数，先按点数排序，再按花色排序
func sortHand(hand []Card) {
	sort.Slice(hand, func(i, j int) bool {
		// 处理大小王的情况
		if isJoker(hand[i].Rank) {
			return false
		}
		if isJoker(hand[j].Rank) {
			return true
		}
		// 首先根据点数排序
		if rankOrder[hand[i].Rank] != rankOrder[hand[j].Rank] {
			return rankOrder[hand[i].Rank] < rankOrder[hand[j].Rank]
		}
		// 如果点数相同，根据花色排序
		return suitOrder[hand[i].Suit] < suitOrder[hand[j].Suit]
	})
}

// 判断是否为大小王
func isJoker(rank string) bool {
	for _, joker := range jokers {
		if rank == joker {
			return true
		}
	}
	return false
}

// 发牌函数，发完所有牌
func dealAllCards(deck *[]Card, numPlayers int) ([][]Card, error) {
	cardsPerPlayer := len(*deck) / numPlayers
	playersHands := make([][]Card, numPlayers)

	for i := 0; i < cardsPerPlayer; i++ {
		for j := 0; j < numPlayers; j++ {
			playersHands[j] = append(playersHands[j], (*deck)[0])
			*deck = (*deck)[1:] // 从牌堆中移除已发出的牌
		}
	}

	// 如果还有剩余的牌，继续分发
	if len(*deck) > 0 {
		for i := 0; len(*deck) > 0; i = (i + 1) % numPlayers {
			playersHands[i] = append(playersHands[i], (*deck)[0])
			*deck = (*deck)[1:]
		}
	}

	return playersHands, nil
}

func main() {
	// 控制是否包含大小王
	includeJokers := true

	// 创建一副牌并洗牌
	deck := createDeck(includeJokers)
	shuffle(deck)

	// 假设给 4 个玩家发所有的牌
	numPlayers := 4

	players, err := dealAllCards(&deck, numPlayers)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 对每个玩家的手牌进行排序
	for i := range players {
		sortHand(players[i])
	}

	// 输出每个玩家的手牌（符号格式）
	for i, hand := range players {
		fmt.Printf("玩家 %d 的手牌: %v\n", i+1, hand)
	}

	// 检查是否所有牌都发完
	fmt.Printf("剩余牌的数量: %d 张\n", len(deck))
}
