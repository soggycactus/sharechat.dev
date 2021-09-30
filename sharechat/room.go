package sharechat

import (
	"log"

	"github.com/google/uuid"
)

type RoomRepository interface {
	InsertRoom(room *Room) error
	GetRoom(roomID string) (*Room, error)
	GetRoomMembers(roomID string) (*[]Member, error)
	DeleteRoom(roomID string) error
}

type Room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// outbound notifies all Online members of a new message
	outbound chan Message
	// subscribe adds a member to the Room and sets their status to Online
	subscribe chan Member
	// leave removes a member from the room entirely
	leave chan Member
	// Members holds all the local members of a room
	members     map[string]Member
	memberRepo  MemberRepository
	messageRepo MessageRepository
}

func NewRoom(name string, memberRepo MemberRepository, messageRepo MessageRepository) *Room {
	return &Room{
		ID:          uuid.New().String(),
		Name:        name,
		outbound:    make(chan Message),
		subscribe:   make(chan Member),
		leave:       make(chan Member),
		members:     make(map[string]Member),
		memberRepo:  memberRepo,
		messageRepo: messageRepo,
	}
}

func (r *Room) Members() map[string]Member {
	return r.members
}

// Room constantly listens for messages and forwards them to existing
func (r *Room) Start() {
	log.Printf("starting room %v", r.ID)
	for {
		select {
		case message := <-r.outbound:
			if err := r.messageRepo.InsertMessage(message); err == nil {
				for _, member := range r.members {
					r.notifyMember(message, member)
				}
			} else {
				log.Printf("failed to insert message in room %s: %v", r.ID, err)
			}
		case newMember := <-r.subscribe:
			{
				if message, err := r.memberRepo.InsertMember(newMember); err == nil {
					r.members[newMember.ID] = newMember
					for _, member := range r.members {
						r.notifyMember(
							*message,
							member,
						)
					}
				} else {
					log.Printf("failed to insert member %s in room %s: %v", newMember.Name, r.ID, err)
				}
			}
		case leaveMember := <-r.leave:
			{
				if message, err := r.memberRepo.DeleteMember(leaveMember); err == nil {
					delete(r.members, leaveMember.ID)
					for _, member := range r.members {
						r.notifyMember(
							*message,
							member,
						)
					}
				} else {
					log.Printf("failed to remove member %s from room %s: %v", leaveMember.Name, r.ID, err)
				}
			}
		}
	}
}

func (r *Room) notifyMember(message Message, member Member) {
	select {
	case member.inbound <- message:
	// if the send blocks, remove them from the room
	default:
		if message, err := r.memberRepo.DeleteMember(member); err != nil {
			log.Printf("failed to remove member %s from room %s: %v", member.Name, r.ID, err)
		} else {
			delete(r.members, member.ID)
			for _, member := range r.members {
				r.notifyMember(
					*message,
					member,
				)
			}
		}
	}
}
