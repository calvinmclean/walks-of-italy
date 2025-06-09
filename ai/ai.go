package ai

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
	"walks-of-italy/ai/tools"
	"walks-of-italy/storage"

	"github.com/ollama/ollama/api"
)

func Chat(sc *storage.Client, model, accessToken string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	printer := make(chan string)
	go func() {
		for c := range printer {
			fmt.Print(c)
		}
	}()

	walksTools := tools.New(sc, accessToken, *slog.Default())

	allTours, _ := walksTools.GetAllTours(tools.GetAllToursInput{})

	ch := &chatHandler{
		printer: printer,
		messages: []api.Message{
			{
				Role:    "system",
				Content: fmt.Sprintf("The current time is %s", time.Now().Format(time.RFC1123)),
			},
			{
				Role:    "system",
				Content: "Here are the tours: " + allTours,
			},
		},
		client: client,
		walks:  walksTools,
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">>> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()

		ch.messages = append(ch.messages, api.Message{
			Role:    "user",
			Content: line,
		})

		err := ch.doChat(model)
		if err != nil {
			return err
		}
		fmt.Println()
	}

	return nil
}

func (ch *chatHandler) doChat(model string) error {
	// reset chat-scoped bools
	ch.done = false
	ch.usingTool = false

	for !ch.done {
		ctx := context.Background()
		req := &api.ChatRequest{
			Model:    model,
			Messages: ch.messages,
			Tools:    ch.walks.Tools(),
		}

		err := ch.client.Chat(ctx, req, ch.handleResponse)
		if err != nil {
			return err
		}
	}

	return nil
}

type chatHandler struct {
	client   *api.Client
	printer  chan<- string
	messages []api.Message
	walks    tools.Tools

	usingTool bool
	done      bool
}

func (ch *chatHandler) handleResponse(resp api.ChatResponse) error {
	if len(resp.Message.ToolCalls) > 0 {
		ch.usingTool = true

		tc := resp.Message.ToolCalls[0].Function
		content, err := ch.walks.Execute(tc.Name, tc.Arguments)
		if err != nil {
			return err
		}

		ch.messages = append(ch.messages, api.Message{
			Role:    "tool",
			Content: content,
		})
		return nil
	}

	ch.printer <- resp.Message.Content

	// After receiving the ToolCall response, the model will respond "done", but still needs to
	// answer the prompt
	if resp.Done {
		if ch.usingTool {
			ch.usingTool = false
			return nil
		}
		ch.done = true
	}

	return nil
}
