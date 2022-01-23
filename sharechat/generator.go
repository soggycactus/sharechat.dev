package sharechat

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Generator struct {
	mu         *sync.Mutex
	adjectives []string
	colors     []string
	nouns      []string
	random     *rand.Rand
}

func (g *Generator) GenerateRoomName() string {
	g.mu.Lock()
	adjective := g.adjectives[g.random.Intn(len(g.adjectives))]
	noun := g.nouns[g.random.Intn(len(g.nouns))]
	g.mu.Unlock()

	return fmt.Sprintf("%v %v", strings.Title(adjective), strings.Title(noun))
}

func (g *Generator) GenerateMemberName() string {
	g.mu.Lock()
	color := g.colors[g.random.Intn(len(g.colors))]
	noun := g.nouns[g.random.Intn(len(g.nouns))]
	g.mu.Unlock()

	return fmt.Sprintf("%v %v", strings.Title(color), strings.Title(noun))
}

func NewGenerator() Generator {
	generator := &Generator{
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
		mu:         new(sync.Mutex),
		adjectives: ADJECTIVES,
		colors:     COLORS,
		nouns:      NOUNS,
	}

	return *generator
}
