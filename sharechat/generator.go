package sharechat

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Generator struct {
	mu         *sync.Mutex
	adjectives []string
	colors     []string
	nouns      []string
	random     *rand.Rand
	caser      *cases.Caser
}

func (g *Generator) GenerateRoomName() string {
	g.mu.Lock()
	adjective := g.adjectives[g.random.Intn(len(g.adjectives))]
	noun := g.nouns[g.random.Intn(len(g.nouns))]
	g.mu.Unlock()

	return fmt.Sprintf("%v %v", g.caser.String(adjective), g.caser.String(noun))
}

func (g *Generator) GenerateMemberName() string {
	g.mu.Lock()
	color := g.colors[g.random.Intn(len(g.colors))]
	noun := g.nouns[g.random.Intn(len(g.nouns))]
	g.mu.Unlock()

	return fmt.Sprintf("%v %v", g.caser.String(color), g.caser.String(noun))
}

func NewGenerator() Generator {
	caser := cases.Title(language.AmericanEnglish)
	generator := &Generator{
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
		mu:         new(sync.Mutex),
		adjectives: ADJECTIVES,
		colors:     COLORS,
		nouns:      NOUNS,
		caser:      &caser,
	}

	return *generator
}
