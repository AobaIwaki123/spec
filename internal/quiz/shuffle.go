package quiz

import (
	"math/rand"
	"time"
)

// ShuffledQuestion wraps Question with dynamically randomized options.
type ShuffledQuestion struct {
	Original        Question
	ShuffledOptions []string
	NewAnswerIndex  int
	IndexMap        []int // IndexMap[shuffledIndex] = originalIndex
}

// ShuffleQuestion randomizes options while maintaining accurate answer index mapping.
func ShuffleQuestion(q Question) ShuffledQuestion {
	n := len(q.Options)
	if n <= 1 {
		return ShuffledQuestion{
			Original:        q,
			ShuffledOptions: q.Options,
			NewAnswerIndex:  q.AnswerIndex,
			IndexMap:        []int{0},
		}
	}

	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(n, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	shuffledOpts := make([]string, n)
	newAnswerIndex := -1
	for newIdx, origIdx := range indices {
		shuffledOpts[newIdx] = q.Options[origIdx]
		if origIdx == q.AnswerIndex {
			newAnswerIndex = newIdx
		}
	}

	return ShuffledQuestion{
		Original:        q,
		ShuffledOptions: shuffledOpts,
		NewAnswerIndex:  newAnswerIndex,
		IndexMap:        indices,
	}
}

// ResolveOriginalChoice converts user choice index on shuffled options back to original option index.
func (sq *ShuffledQuestion) ResolveOriginalChoice(userChoice int) int {
	if userChoice < 0 || userChoice >= len(sq.IndexMap) {
		return -1
	}
	return sq.IndexMap[userChoice]
}
