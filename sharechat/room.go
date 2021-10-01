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
	members     map[string]*Member
	memberRepo  MemberRepository
	messageRepo MessageRepository
	// testNoop is a hook for synchronizing goroutines in tests.
	// It is called at the end of each receive in Start
	testNoop func()
	// flag used in tests to turn off error logging
	logErrors bool
}

func NewRoom(name string, memberRepo MemberRepository, messageRepo MessageRepository) *Room {
	return &Room{
		ID:          uuid.New().String(),
		Name:        name,
		outbound:    make(chan Message),
		subscribe:   make(chan Member),
		leave:       make(chan Member),
		members:     make(map[string]*Member),
		memberRepo:  memberRepo,
		messageRepo: messageRepo,
		testNoop:    func() {},
		logErrors:   true,
	}
}

func (r *Room) Members() map[string]*Member {
	return r.members
}

// Room constantly listens for messages and forwards them to existing members
func (r *Room) Start() {
	log.Printf("starting room %v", r.ID)
	for {
		select {
		case message := <-r.outbound:
			if err := r.messageRepo.InsertMessage(message); err == nil {
				for _, member := range r.members {
					r.notifyMember(message, *member)
				}
			} else {
				if r.logErrors {
					log.Printf("failed to insert message %s in room %s: %v", message.ID, r.ID, err)
				}
			}
		case newMember := <-r.subscribe:
			if message, err := r.memberRepo.InsertMember(newMember); err == nil {
				r.members[newMember.ID] = &newMember
				for _, member := range r.members {
					r.notifyMember(
						*message,
						*member,
					)
				}
			} else {
				if r.logErrors {
					log.Printf("failed to insert member %s in room %s: %v", newMember.Name, r.ID, err)
				}
			}
		case leaveMember := <-r.leave:
			if message, err := r.memberRepo.DeleteMember(leaveMember); err == nil {
				delete(r.members, leaveMember.ID)
				for _, member := range r.members {
					r.notifyMember(
						*message,
						*member,
					)
				}
			} else {
				if r.logErrors {
					log.Printf("failed to remove member %s from room %s: %v", leaveMember.Name, r.ID, err)
				}
			}
		}
		r.testNoop()
	}
}

func (r *Room) Subscribe(member Member) {
	r.subscribe <- member
}

func (r *Room) Outbound(message Message) {
	r.outbound <- message
}

func (r *Room) Leave(member Member) {
	r.leave <- member
}

func (r *Room) notifyMember(message Message, member Member) {
	select {
	case member.inbound <- message:
	// if the send blocks, log it an move on
	default:
		if r.logErrors {
			log.Printf("failed to notify member %s of message %s", member.ID, message.ID)
		}
	}
}

func (r *Room) WithMembers(member ...Member) *Room {
	for _, mem := range member {
		r.members[mem.ID] = &mem
	}
	return r
}

func (r *Room) WithTestNoop(f func()) *Room {
	r.testNoop = f
	return r
}

func (r *Room) WithNoErrorLogs() *Room {
	r.logErrors = false
	return r
}
