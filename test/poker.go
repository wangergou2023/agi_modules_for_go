package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// 定义扑克牌的花色符号和点数
var suits = []string{"♥", "♦", "♣", "♠"}
var ranks = []string{"3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A", "2"}

// 定义大小王
var jokers = []string{"🃏大", "🃏小"}

// 定义花色和点数的优先级，用于排序
var suitOrder = map[string]int{"♥": 1, "♦": 2, "♣": 3, "♠": 4}
var rankOrder = map[string]int{
	"3": 1, "4": 2, "5": 3, "6": 4, "7": 5, "8": 6, "9": 7, "10": 8, "J": 9, "Q": 10, "K": 11, "A": 12, "2": 13,
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

// 发牌函数，给三位玩家发17张牌，留下3张底牌
func dealCardsForDouDiZhu(deck *[]Card) ([][]Card, []Card) {
	numPlayers := 3
	playersHands := make([][]Card, numPlayers)

	// 发每位玩家17张牌
	for i := 0; i < 17; i++ {
		for j := 0; j < numPlayers; j++ {
			playersHands[j] = append(playersHands[j], (*deck)[0])
			*deck = (*deck)[1:] // 从牌堆中移除已发出的牌
		}
	}

	// 剩下3张底牌
	bottomCards := (*deck)[:3]
	*deck = (*deck)[3:]

	return playersHands, bottomCards
}

// 玩家出牌函数，记录每位玩家出牌
func playCard(playerHand *[]Card, cardIndex int, tracker []Card) Card {
	card := (*playerHand)[cardIndex]
	*playerHand = append((*playerHand)[:cardIndex], (*playerHand)[cardIndex+1:]...)
	tracker = append(tracker, card) // 记录该牌已被出
	return card
}

// 排序已出牌的情况
func sortPlayedCards(cards []Card) {
	sortHand(cards) // 直接使用已有的排序函数
}

func main() {
	// 控制是否包含大小王
	includeJokers := true

	// 创建一副牌并洗牌
	deck := createDeck(includeJokers)
	shuffle(deck)

	// 发牌并留底牌
	players, bottomCards := dealCardsForDouDiZhu(&deck)

	// 假设第一个玩家为地主
	landlordIndex := 0
	players[landlordIndex] = append(players[landlordIndex], bottomCards...)

	// 对每个玩家的手牌进行排序
	for i := range players {
		sortHand(players[i])
	}

	// 记录出牌的情况
	var cardTracker []Card

	// 输出每个玩家的手牌（符号格式）
	for i, hand := range players {
		if i == landlordIndex {
			fmt.Printf("玩家 %d (地主) 的手牌: %v\n", i+1, hand)
		} else {
			fmt.Printf("玩家 %d (农民) 的手牌: %v\n", i+1, hand)
		}
	}

	// 输出底牌
	fmt.Printf("底牌: %v\n\n", bottomCards)

	// 模拟出牌过程，假设每个玩家出第一张牌
	fmt.Println("出牌记录:")
	for round := 0; round < 2; round++ { // 模拟出牌
		for i := range players {
			if len(players[i]) > 0 {
				card := playCard(&players[i], 0, cardTracker) // 每个玩家出第一张牌
				cardTracker = append(cardTracker, card)       // 记录已出牌
				fmt.Printf("玩家 %d 出牌: %s\n", i+1, card.String())
			}
		}
	}

	// 对已出牌的牌进行排序
	sortPlayedCards(cardTracker)

	// 输出已出牌的情况
	fmt.Printf("\n已出牌的牌 (排序后):%v\n", cardTracker)

	// 输出剩余手牌
	fmt.Println("\n剩余手牌:")
	for i, hand := range players {
		if i == landlordIndex {
			fmt.Printf("玩家 %d (地主) 的手牌: %v\n", i+1, hand)
		} else {
			fmt.Printf("玩家 %d (农民) 的手牌: %v\n", i+1, hand)
		}
	}
}
