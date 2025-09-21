package logic

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"db-desktop/backend/utils"
)

// ConfirmCard represents a confirmation card for MCP tool execution
type ConfirmCard struct {
	CardID          string    `json:"cardId"`      // UUID generated card ID
	ShowContent     string    `json:"showContent"` // Content to display in the card
	ConfirmCallback func()    `json:"-"`           // Function to call when confirmed
	RejectCallback  func()    `json:"-"`           // Function to call when rejected
	CreatedAt       time.Time `json:"createdAt"`   // When the card was created
	Status          string    `json:"status"`      // pending, confirmed, rejected, expired
	ExpiresAt       time.Time `json:"expiresAt"`   // When the card expires
	// 额外的元数据
	ConversationID string `json:"conversationId,omitempty"` // 对话ID
	ToolCallID     string `json:"toolCallId,omitempty"`     // 工具调用ID
}

// CardManager manages confirmation cards
type CardManager struct {
	cards map[string]*ConfirmCard
	mu    sync.RWMutex
}

// generateCardID generates a unique card ID
func (cm *CardManager) generateCardID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// CreateCard creates a new confirmation card
func (cm *CardManager) CreateCard(showContent string, confirmCallback, rejectCallback func()) *ConfirmCard {
	return cm.CreateCardWithMetadata(showContent, confirmCallback, rejectCallback, "", "")
}

// CreateCardWithMetadata creates a new confirmation card with additional metadata
func (cm *CardManager) CreateCardWithMetadata(showContent string, confirmCallback, rejectCallback func(), conversationID, toolCallID string) *ConfirmCard {
	cardID := cm.generateCardID()
	now := time.Now()

	card := &ConfirmCard{
		CardID:          cardID,
		ShowContent:     showContent,
		ConfirmCallback: confirmCallback,
		RejectCallback:  rejectCallback,
		CreatedAt:       now,
		Status:          "pending",
		ExpiresAt:       now.Add(5 * time.Minute), // Cards expire after 5 minutes
		ConversationID:  conversationID,
		ToolCallID:      toolCallID,
	}

	cm.mu.Lock()
	cm.cards[cardID] = card
	cm.mu.Unlock()

	utils.Infof("Created confirmation card with metadata: cardID=%s, showContent=%s, conversationID=%s, toolCallID=%s, expiresAt=%s", cardID, showContent, conversationID, toolCallID, card.ExpiresAt)

	// Start expiration timer
	go cm.startExpirationTimer(cardID)

	return card
}

// GetCard retrieves a card by ID
func (cm *CardManager) GetCard(cardID string) (*ConfirmCard, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	card, exists := cm.cards[cardID]
	return card, exists
}

// GetAllCards returns all cards
func (cm *CardManager) GetAllCards() []*ConfirmCard {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*ConfirmCard
	for _, card := range cm.cards {
		result = append(result, card)
	}
	return result
}

// GetPendingCards returns all pending cards
func (cm *CardManager) GetPendingCards() []*ConfirmCard {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*ConfirmCard
	for _, card := range cm.cards {
		if card.Status == "pending" {
			result = append(result, card)
		}
	}
	return result
}

// ConfirmCard confirms a card and executes the confirm callback
func (cm *CardManager) ConfirmCard(cardID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	utils.Infof("🔍 Looking up card for confirmation: cardID=%s", cardID)

	card, exists := cm.cards[cardID]
	if !exists {
		utils.Errorf("Card not found for confirmation: cardID=%s", cardID)
		return fmt.Errorf("card not found: %s", cardID)
	}

	utils.Infof("📋 Card found, checking status: cardID=%s, currentStatus=%s, conversationID=%s, toolCallID=%s, expiresAt=%s", cardID, card.Status, card.ConversationID, card.ToolCallID, card.ExpiresAt)

	if card.Status != "pending" {
		utils.Errorf("Card already processed: cardID=%s, currentStatus=%s", cardID, card.Status)
		return fmt.Errorf("card already processed: %s (status: %s)", cardID, card.Status)
	}

	// Check if card has expired
	if time.Now().After(card.ExpiresAt) {
		card.Status = "expired"
		utils.Errorf("Card has expired: cardID=%s, expiresAt=%s", cardID, card.ExpiresAt)
		return fmt.Errorf("card has expired: %s", cardID)
	}

	// Update status
	card.Status = "confirmed"

	utils.Infof("✅ Card confirmed, executing callback: cardID=%s, conversationID=%s, toolCallID=%s", cardID, card.ConversationID, card.ToolCallID)

	// Execute confirm callback in a goroutine to avoid blocking
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Errorf("Panic in confirm callback: cardID=%s, panic=%s", cardID, r)
			}
		}()

		utils.Infof("🚀 Executing confirm callback: cardID=%s", cardID)

		if card.ConfirmCallback != nil {
			card.ConfirmCallback()
			utils.Infof("✅ Confirm callback executed successfully: cardID=%s", cardID)
		} else {
			utils.Warnf("No confirm callback set for card: cardID=%s", cardID)
		}
	}()

	return nil
}

// RejectCard rejects a card and executes the reject callback
func (cm *CardManager) RejectCard(cardID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	utils.Infof("🔍 Looking up card for rejection: cardID=%s", cardID)

	card, exists := cm.cards[cardID]
	if !exists {
		utils.Errorf("Card not found for rejection: cardID=%s", cardID)
		return fmt.Errorf("card not found: %s", cardID)
	}

	utils.Infof("📋 Card found, checking status: cardID=%s, currentStatus=%s, conversationID=%s, toolCallID=%s, expiresAt=%s", cardID, card.Status, card.ConversationID, card.ToolCallID, card.ExpiresAt)

	if card.Status != "pending" {
		utils.Errorf("Card already processed: cardID=%s, currentStatus=%s", cardID, card.Status)
		return fmt.Errorf("card already processed: %s (status: %s)", cardID, card.Status)
	}

	// Check if card has expired
	if time.Now().After(card.ExpiresAt) {
		card.Status = "expired"
		utils.Errorf("Card has expired: cardID=%s, expiresAt=%s", cardID, card.ExpiresAt)
		return fmt.Errorf("card has expired: %s", cardID)
	}

	// Update status
	card.Status = "rejected"

	utils.Infof("❌ Card rejected, executing callback: cardID=%s, conversationID=%s, toolCallID=%s", cardID, card.ConversationID, card.ToolCallID)

	// Execute reject callback in a goroutine to avoid blocking
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Errorf("Panic in reject callback: cardID=%s, panic=%s", cardID, r)
			}
		}()

		utils.Infof("🚀 Executing reject callback: cardID=%s", cardID)

		if card.RejectCallback != nil {
			card.RejectCallback()
			utils.Infof("✅ Reject callback executed successfully: cardID=%s", cardID)
		} else {
			utils.Warnf("No reject callback set for card: cardID=%s", cardID)
		}
	}()

	return nil
}

// RemoveCard removes a card from the manager
func (cm *CardManager) RemoveCard(cardID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.cards, cardID)
	utils.Infof("Card removed - cardID: %s", cardID)
}

// CleanupExpiredCards removes expired cards
func (cm *CardManager) CleanupExpiredCards() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var expiredCards []string

	for cardID, card := range cm.cards {
		if now.After(card.ExpiresAt) && card.Status == "pending" {
			card.Status = "expired"
			expiredCards = append(expiredCards, cardID)
		}
	}

	// Remove expired cards
	for _, cardID := range expiredCards {
		delete(cm.cards, cardID)
	}

	if len(expiredCards) > 0 {
		utils.Infof("Cleaned up expired cards - count: %d", len(expiredCards))
	}
}

// startExpirationTimer starts a timer to check for card expiration
func (cm *CardManager) startExpirationTimer(cardID string) {
	cm.mu.RLock()
	card, exists := cm.cards[cardID]
	if !exists {
		cm.mu.RUnlock()
		return
	}
	expiresAt := card.ExpiresAt
	cm.mu.RUnlock()

	// Calculate duration until expiration
	duration := time.Until(expiresAt)
	if duration <= 0 {
		// Card already expired
		cm.CleanupExpiredCards()
		return
	}

	// Wait for expiration
	time.Sleep(duration)

	// Check if card still exists and is pending
	cm.mu.RLock()
	card, exists = cm.cards[cardID]
	if exists && card.Status == "pending" {
		cm.mu.RUnlock()
		cm.CleanupExpiredCards()
	} else {
		cm.mu.RUnlock()
	}
}

// GetCardStats returns statistics about cards
func (cm *CardManager) GetCardStats() map[string]int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := map[string]int{
		"total":     0,
		"pending":   0,
		"confirmed": 0,
		"rejected":  0,
		"expired":   0,
	}

	for _, card := range cm.cards {
		stats["total"]++
		stats[card.Status]++
	}

	return stats
}
