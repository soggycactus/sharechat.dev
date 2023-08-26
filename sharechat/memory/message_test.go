//go:build unit || all

package memory_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/soggycactus/sharechat.dev/sharechat"
	"github.com/soggycactus/sharechat.dev/sharechat/memory"
	"github.com/stretchr/testify/assert"
)

func TestMessageRepo(t *testing.T) {
	repo := memory.NewMessageRepo()

	member := sharechat.NewMember("test", "test", nil)

	insertedMessages := []sharechat.Message{}
	for i := 0; i < 10; i++ {
		message, err := repo.InsertMessage(context.TODO(), sharechat.NewChatMessage(*member, []byte(fmt.Sprintf("test %d", i))))
		if err != nil {
			t.Fatal(err)
		}

		insertedMessages = append(insertedMessages, *message)
	}

	queriedMessages := []sharechat.Message{
		insertedMessages[0],
	}
	cursor := sharechat.MessageCursor{
		ID:   insertedMessages[0].ID,
		Sent: insertedMessages[0].Sent,
	}
	for {
		messages, err := repo.GetMessages(context.TODO(), sharechat.GetMessageOptions{
			Limit: 2,
			After: cursor,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(messages) == 0 {
			break
		}

		queriedMessages = append(queriedMessages, messages...)
		cursor = sharechat.MessageCursor{
			ID:   messages[len(messages)-1].ID,
			Sent: messages[len(messages)-1].Sent,
		}
	}

	sort.Slice(queriedMessages, func(i, j int) bool {
		return queriedMessages[i].Sent.Before(queriedMessages[j].Sent)
	})

	assert.Len(t, queriedMessages, 10, "should have all messages")
	assert.Equal(t, insertedMessages, queriedMessages, "should return the same messages")

	queriedMessages = []sharechat.Message{insertedMessages[len(insertedMessages)-1]}
	cursor = sharechat.MessageCursor{
		ID:   insertedMessages[len(insertedMessages)-1].ID,
		Sent: insertedMessages[len(insertedMessages)-1].Sent,
	}
	for {
		messages, err := repo.GetMessages(context.TODO(), sharechat.GetMessageOptions{
			Limit:  2,
			Before: cursor,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(messages) == 0 {
			break
		}

		queriedMessages = append(queriedMessages, messages...)
		cursor = sharechat.MessageCursor{
			ID:   messages[len(messages)-1].ID,
			Sent: messages[len(messages)-1].Sent,
		}
	}

	sort.Slice(queriedMessages, func(i, j int) bool {
		return queriedMessages[i].Sent.Before(queriedMessages[j].Sent)
	})

	assert.Len(t, queriedMessages, 10, "should have all messages")
	assert.Equal(t, insertedMessages, queriedMessages, "should return the same messages")

}
