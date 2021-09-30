package memory

import (
	"errors"

	"github.com/soggycactus/sharechat.dev/sharechat"
)

type MemberRepo struct {
	Members     map[string]sharechat.Member
	messageRepo sharechat.MessageRepository
}

func NewMemberRepo(messageRepo sharechat.MessageRepository) MemberRepo {
	return MemberRepo{
		Members:     make(map[string]sharechat.Member),
		messageRepo: messageRepo,
	}
}

func (m *MemberRepo) InsertMember(member sharechat.Member) (*sharechat.Message, error) {
	m.Members[member.ID] = member
	message := sharechat.NewMemberJoinedMessage(member)
	if err := m.messageRepo.InsertMessage(message); err != nil {
		delete(m.Members, member.ID)
		return nil, errors.New("failed to insert member")
	}
	return &message, nil
}

func (m *MemberRepo) GetMembersByRoom(roomID string) (*[]sharechat.Member, error) {
	members := []sharechat.Member{}

	for _, member := range m.Members {
		if member.RoomID() == roomID {
			members = append(members, member)
		}
	}

	return &members, nil
}

func (m *MemberRepo) DeleteMember(member sharechat.Member) (*sharechat.Message, error) {
	message := sharechat.NewMemberLeftMessage(member)
	if err := m.messageRepo.InsertMessage(message); err != nil {
		return nil, errors.New("failed to delete member")
	}
	delete(m.Members, member.ID)

	return &message, nil
}
