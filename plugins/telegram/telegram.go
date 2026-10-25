package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/5c077m4n/harel/plugin"
)

type Plugin struct {
	token    string
	offset   int64
	lastChat int64
	client   *http.Client
}

func New() plugin.Plugin {
	return &Plugin{
		token:  os.Getenv("TELEGRAM_BOT_TOKEN"),
		client: &http.Client{Timeout: 35 * time.Second},
	}
}

func (p *Plugin) Name() string {
	return "telegram"
}

func (p *Plugin) Start(ctx context.Context, in chan<- string, out <-chan string) error {
	if p.token == "" {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case reply := <-out:
			if err := p.send(p.lastChat, reply); err != nil {
				return err
			}
		default:
		}
		updates, err := p.poll(ctx)
		if err != nil {
			return err
		}
		for _, u := range updates {
			if u.UpdateID >= p.offset {
				p.offset = u.UpdateID + 1
			}
			if u.Message.Text != "" {
				p.lastChat = u.Message.Chat.ID
				select {
				case in <- u.Message.Text:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}
}

type update struct {
	UpdateID int64 `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func (p *Plugin) poll(ctx context.Context) ([]update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&limit=100&timeout=5", p.token, p.offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("telegram getUpdates: %d %s", resp.StatusCode, b)
	}
	var result struct {
		Ok     bool     `json:"ok"`
		Result []update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if !result.Ok {
		return nil, fmt.Errorf("telegram poll failed")
	}
	return result.Result, nil
}

func (p *Plugin) send(chatID int64, text string) error {
	body, err := json.Marshal(map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", p.token), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("telegram sendMessage: %d %s", resp.StatusCode, b)
	}
	return nil
}
