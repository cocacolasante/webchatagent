package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const cachePrefix = "blueprint-chat:prompt-cache:"

// Assembler builds system prompts from tenant knowledge.
type Assembler struct {
	redis    *redis.Client
	cacheTTL time.Duration
}

// NewAssembler creates a new Assembler.
func NewAssembler(redisClient *redis.Client, cacheTTL time.Duration) *Assembler {
	return &Assembler{redis: redisClient, cacheTTL: cacheTTL}
}

// BuildSystemPrompt assembles the full system prompt for a tenant.
// Results are cached in Redis for cacheTTL duration.
func (a *Assembler) BuildSystemPrompt(ctx context.Context, tenantID string, bizInfo BusinessInfo, kb KnowledgeBase) (string, error) {
	cacheKey := cachePrefix + tenantID

	if cached, err := a.redis.Get(ctx, cacheKey).Result(); err == nil {
		return cached, nil
	}

	prompt := a.assemble(bizInfo, kb)

	if err := a.redis.Set(ctx, cacheKey, prompt, a.cacheTTL).Err(); err != nil {
		// Log but don't fail — cache miss is acceptable
		_ = err
	}

	return prompt, nil
}

// InvalidateCache removes the cached system prompt for a tenant.
func (a *Assembler) InvalidateCache(ctx context.Context, tenantID string) error {
	return a.redis.Del(ctx, cachePrefix+tenantID).Err()
}

func (a *Assembler) assemble(bizInfo BusinessInfo, kb KnowledgeBase) string {
	var sb strings.Builder

	sb.WriteString("You are a helpful AI assistant for a business. Be professional, warm, and concise.\n\n")

	// Business context
	sb.WriteString("## Business Information\n")
	if bizInfo.Name != "" {
		sb.WriteString(fmt.Sprintf("Business Name: %s\n", bizInfo.Name))
	}
	if bizInfo.Tagline != "" {
		sb.WriteString(fmt.Sprintf("Tagline: %s\n", bizInfo.Tagline))
	}
	if bizInfo.Industry != "" {
		sb.WriteString(fmt.Sprintf("Industry: %s\n", bizInfo.Industry))
	}
	if bizInfo.Website != "" {
		sb.WriteString(fmt.Sprintf("Website: %s\n", bizInfo.Website))
	}
	if bizInfo.Email != "" {
		sb.WriteString(fmt.Sprintf("Contact Email: %s\n", bizInfo.Email))
	}
	if bizInfo.Phone != "" {
		sb.WriteString(fmt.Sprintf("Phone: %s\n", bizInfo.Phone))
	}
	if bizInfo.Address != "" {
		sb.WriteString(fmt.Sprintf("Address: %s\n", bizInfo.Address))
	}
	if bizInfo.Hours != "" {
		sb.WriteString(fmt.Sprintf("Hours: %s\n", bizInfo.Hours))
	}

	// Services
	if len(bizInfo.Services) > 0 {
		sb.WriteString("\n## Services Offered\n")
		for _, svc := range bizInfo.Services {
			sb.WriteString(fmt.Sprintf("- **%s**", svc.Name))
			if svc.Price != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", svc.Price))
			}
			if svc.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", svc.Description))
			}
			sb.WriteString("\n")
		}
	}

	// Team
	if len(bizInfo.Team) > 0 {
		sb.WriteString("\n## Team\n")
		for _, member := range bizInfo.Team {
			sb.WriteString(fmt.Sprintf("- %s, %s\n", member.Name, member.Role))
		}
	}

	// FAQs
	if len(kb.FAQs) > 0 {
		sb.WriteString("\n## Frequently Asked Questions\n")
		for _, faq := range kb.FAQs {
			sb.WriteString(fmt.Sprintf("Q: %s\nA: %s\n\n", faq.Question, faq.Answer))
		}
	}

	// Documents
	if len(kb.Documents) > 0 {
		sb.WriteString("\n## Additional Knowledge\n")
		for _, doc := range kb.Documents {
			if doc.Title != "" {
				sb.WriteString(fmt.Sprintf("### %s\n", doc.Title))
			}
			if doc.URL != "" {
				sb.WriteString(fmt.Sprintf("Source: %s\n", doc.URL))
			}
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	// Custom instructions
	if kb.CustomInstructions != "" {
		sb.WriteString("\n## Behavioral Instructions\n")
		sb.WriteString(kb.CustomInstructions)
		sb.WriteString("\n")
	}

	sb.WriteString("\n## General Guidelines\n")
	sb.WriteString("- Keep responses concise and helpful.\n")
	sb.WriteString("- If you don't know something, say so honestly and offer to connect the user with the team.\n")
	sb.WriteString("- Do not make up information about the business.\n")
	sb.WriteString("- If a user expresses interest in a service or wants pricing, gently offer to capture their contact information.\n")
	sb.WriteString("- If a user wants to schedule a meeting or get a demo, offer to show available booking times.\n")

	return sb.String()
}

// ParseBusinessInfo deserializes business info from a JSON map (as stored in DB JSONB).
func ParseBusinessInfo(data []byte) (BusinessInfo, error) {
	var info BusinessInfo
	if len(data) == 0 || string(data) == "null" || string(data) == "{}" {
		return info, nil
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return info, fmt.Errorf("parse business info: %w", err)
	}
	return info, nil
}

// ParseKnowledgeBase deserializes knowledge base from a JSON map.
func ParseKnowledgeBase(data []byte) (KnowledgeBase, error) {
	var kb KnowledgeBase
	if len(data) == 0 || string(data) == "null" || string(data) == "{}" {
		return kb, nil
	}
	if err := json.Unmarshal(data, &kb); err != nil {
		return kb, fmt.Errorf("parse knowledge base: %w", err)
	}
	return kb, nil
}
