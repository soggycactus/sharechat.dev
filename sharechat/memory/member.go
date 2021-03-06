package memory

import (
	"context"
	"errors"
	"time"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type MemberRepo struct {
	Members     map[string]sharechat.Member
	messageRepo sharechat.MessageRepository
}

func NewMemberRepo(messageRepo sharechat.MessageRepository) *MemberRepo {
	return &MemberRepo{
		Members:     make(map[string]sharechat.Member),
		messageRepo: messageRepo,
	}
}

func (m *MemberRepo) InsertMember(ctx context.Context, member sharechat.Member) (*sharechat.Message, error) {
	m.Members[member.ID] = member
	message := sharechat.NewMemberJoinedMessage(member)
	message.Sent = time.Now()
	result, err := m.messageRepo.InsertMessage(ctx, message)
	if err != nil {
		delete(m.Members, member.ID)
		return nil, errors.New("failed to insert member")
	}
	return result, nil
}

func (m *MemberRepo) GetMembersByRoom(ctx context.Context, roomID string) (*[]sharechat.Member, error) {
	members := []sharechat.Member{}

	for _, member := range m.Members {
		if member.RoomID == roomID {
			members = append(members, member)
		}
	}

	return &members, nil
}

func (m *MemberRepo) DeleteMember(ctx context.Context, member sharechat.Member) (*sharechat.Message, error) {
	message := sharechat.NewMemberLeftMessage(member)
	message.Sent = time.Now()
	result, err := m.messageRepo.InsertMessage(ctx, message)
	if err != nil {
		return nil, errors.New("failed to delete member")
	}
	delete(m.Members, member.ID)

	return result, nil
}
