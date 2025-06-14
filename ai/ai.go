package ai

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"walks-of-italy/ai/tools"
	"walks-of-italy/storage"

	"github.com/ollama/ollama/api"
)

func Chat(sc *storage.Client, model, ventrataToken, walksToken string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	walksTools := tools.New(sc, ventrataToken, walksToken, *slog.Default())

	allTours, _ := walksTools.GetAllTours(tools.GetAllToursInput{})

	ch := &chatHandler{
		messages: []api.Message{
			{
				Role:    "system",
				Content: fmt.Sprintf("The current time is %s", time.Now().Format(time.RFC1123)),
			},
			{
				Role:    "system",
				Content: fmt.Sprintf("Here are the tours: %s. Use this information if the user asks about a specific tour.", allTours),
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
			Think:    new(bool),
		}

		ch.responseMessage = []string{}

		err := ch.client.Chat(ctx, req, ch.handleResponse)
		if err != nil {
			return err
		}

		if len(ch.responseMessage) > 0 {
			content := strings.Join(ch.responseMessage, "")
			ch.messages = append(ch.messages, api.Message{
				Role:    "assistant",
				Content: content,
			})
		}
	}

	return nil
}

type chatHandler struct {
	client   *api.Client
	messages []api.Message
	walks    tools.Tools

	// Collect the streamed response to add to the history after completing the response
	responseMessage []string

	usingTool bool
	done      bool
}

func (ch *chatHandler) handleResponse(resp api.ChatResponse) error {
	if len(resp.Message.ToolCalls) > 0 {
		ch.messages = append(ch.messages, resp.Message)

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

	fmt.Print(resp.Message.Content)
	if len(resp.Message.Content) > 0 {
		ch.responseMessage = append(ch.responseMessage, resp.Message.Content)
	}

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
