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

// 生成记牌器，用于记录已出牌
func createCardTracker() map[string]bool {
	tracker := make(map[string]bool)
	return tracker
}

// 更新记牌器，记录出牌
func updateCardTracker(tracker map[string]bool, card Card) {
	tracker[card.String()] = true
}

// 发牌函数，发完所有牌并根据是否有地主发牌
func dealAllCards(deck *[]Card, numPlayers int, hasLandlord bool, landlordIndex int) ([][]Card, error) {
	cardsPerPlayer := len(*deck) / numPlayers
	playersHands := make([][]Card, numPlayers)

	for i := 0; i < cardsPerPlayer; i++ {
		for j := 0; j < numPlayers; j++ {
			playersHands[j] = append(playersHands[j], (*deck)[0])
			*deck = (*deck)[1:] // 从牌堆中移除已发出的牌
		}
	}

	// 如果有地主且有剩余牌（如斗地主的3张地主牌），将牌给地主玩家
	if hasLandlord && len(*deck) >= 3 {
		playersHands[landlordIndex] = append(playersHands[landlordIndex], (*deck)[:3]...)
		*deck = (*deck)[3:] // 移除地主的牌
	}

	return playersHands, nil
}

// 玩家出牌函数，记录每位玩家出牌
func playCard(playerHand *[]Card, cardIndex int, tracker map[string]bool) Card {
	card := (*playerHand)[cardIndex]
	*playerHand = append((*playerHand)[:cardIndex], (*playerHand)[cardIndex+1:]...)
	updateCardTracker(tracker, card) // 更新记牌器
	return card
}

func main() {
	// 控制是否包含大小王
	includeJokers := true

	// 控制是否有地主
	hasLandlord := true

	// 创建一副牌并洗牌
	deck := createDeck(includeJokers)
	shuffle(deck)

	// 初始化记牌器
	cardTracker := createCardTracker()

	// 假设给 3 个玩家发牌，并通过索引选择地主
	numPlayers := 3
	landlordIndex := 0 // 设定玩家 1 为地主，索引从 0 开始

	players, err := dealAllCards(&deck, numPlayers, hasLandlord, landlordIndex)
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
		if hasLandlord && i == landlordIndex {
			fmt.Printf("玩家 %d (地主) 的手牌: %v\n", i+1, hand)
		} else {
			fmt.Printf("玩家 %d (农民) 的手牌: %v\n", i+1, hand)
		}
	}

	// 模拟出牌过程，假设每个玩家出一张牌
	fmt.Println("出牌记录:")
	for i := range players {
		if len(players[i]) > 0 {
			card := playCard(&players[i], 0, cardTracker) // 每个玩家出第一张牌
			fmt.Printf("玩家 %d 出牌: %s\n", i+1, card.String())
		}
	}

	// 输出已出牌的情况
	fmt.Println("\n记牌器记录已出牌:")
	for card, played := range cardTracker {
		if played {
			fmt.Println(card)
		}
	}

	// 输出每个玩家的手牌（符号格式）
	for i, hand := range players {
		if hasLandlord && i == landlordIndex {
			fmt.Printf("玩家 %d (地主) 的手牌: %v\n", i+1, hand)
		} else {
			fmt.Printf("玩家 %d (农民) 的手牌: %v\n", i+1, hand)
		}
	}

}
