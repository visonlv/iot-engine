package forwarding

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"

	"github.com/visonlv/go-vkit/logger"
)

var (
	ErrInvalidSubject    = errors.New("sublist: invalid subject")
	ErrNotFound          = errors.New("sublist: no matches found")
	ErrNilChan           = errors.New("sublist: nil channel")
	ErrAlreadyRegistered = errors.New("sublist: notification already registered")
)

const (
	pwc = '*'
	fwc = '>'
)

// up.0.pk.sn.property.batch
type subscription struct {
	contextId string
	out       chan *ChanEvent
	ctx       context.Context
	tokens    []string
	from      string
	ruleInfo  *RuleInfo
}

type matchResult struct {
	psubs []*subscription
}

func (r *matchResult) addSubToResult(sub *subscription) *matchResult {
	nr := r.copy()
	nr.psubs = append(nr.psubs, sub)
	return nr
}

func (r *matchResult) copy() *matchResult {
	nr := &matchResult{}
	nr.psubs = append([]*subscription(nil), r.psubs...)
	return nr
}

type node struct {
	next  *level
	psubs map[*subscription]*subscription
}

func newNode() *node {
	return &node{psubs: make(map[*subscription]*subscription)}
}

func (n *node) isEmpty() bool {
	if len(n.psubs) == 0 {
		if n.next == nil || n.next.numNodes() == 0 {
			return true
		}
	}
	return false
}

type level struct {
	nodes    map[string]*node
	pwc, fwc *node
}

func newLevel() *level {
	return &level{nodes: make(map[string]*node)}
}

func (l *level) pruneNode(n *node, t string) {
	if n == nil {
		return
	}
	if n == l.fwc {
		l.fwc = nil
	} else if n == l.pwc {
		l.pwc = nil
	} else {
		delete(l.nodes, t)
	}
}

func (l *level) numNodes() int {
	num := len(l.nodes)
	if l.pwc != nil {
		num++
	}
	if l.fwc != nil {
		num++
	}
	return num
}

type Match struct {
	matches uint64
	inserts uint64
	removes uint64
	count   uint32

	root *level
}

func newMatch() *Match {
	return &Match{
		root: newLevel(),
	}
}

func (s *Match) Insert(sub *subscription) error {
	var sfwc bool
	var n *node
	l := s.root

	for _, t := range sub.tokens {
		lt := len(t)
		if lt == 0 || sfwc {
			return ErrInvalidSubject
		}

		if lt > 1 {
			n = l.nodes[t]
		} else {
			switch t[0] {
			case pwc:
				n = l.pwc
			case fwc:
				n = l.fwc
				sfwc = true
			default:
				n = l.nodes[t]
			}
		}
		if n == nil {
			n = newNode()
			if lt > 1 {
				l.nodes[t] = n
			} else {
				switch t[0] {
				case pwc:
					l.pwc = n
				case fwc:
					l.fwc = n
				default:
					l.nodes[t] = n
				}
			}
		}
		if n.next == nil {
			n.next = newLevel()
		}
		l = n.next
	}

	n.psubs[sub] = sub

	s.count++
	s.inserts++

	s.printInfo("after Insert", sub.tokens, s.root, 1)
	return nil
}

func (s *Match) Match(tokens []string) *matchResult {
	atomic.AddUint64(&s.matches, 1)

	result := &matchResult{}
	s.matchLevel(s.root, tokens, result)

	return result
}

func (s *Match) matchLevel(l *level, toks []string, results *matchResult) {
	var pwc, n *node
	for i, t := range toks {
		if l == nil {
			return
		}
		if l.fwc != nil {
			s.addNodeToResults(l.fwc, results)
		}
		if pwc = l.pwc; pwc != nil {
			s.matchLevel(pwc.next, toks[i+1:], results)
		}
		n = l.nodes[t]
		if n != nil {
			l = n.next
		} else {
			l = nil
		}
	}
	if n != nil {
		s.addNodeToResults(n, results)
	}
	if pwc != nil {
		s.addNodeToResults(pwc, results)
	}
}

func (s *Match) addNodeToResults(n *node, results *matchResult) {
	for _, psub := range n.psubs {
		results.psubs = append(results.psubs, psub)
	}
}

func (s *Match) printInfo(op string, tokens []string, r *level, l int) {
	if len(op) > 0 {
		logger.Infof("op:%s topic:%s", op, strings.Join(tokens, "."))
	}

	nextL := l + 1
	if r.nodes != nil {
		for _, v := range r.nodes {
			for _, v2 := range v.psubs {
				logger.Infof("topic:%s level:%d", strings.Join(v2.tokens, "."), l)
			}
			if v.next != nil {
				s.printInfo("", tokens, v.next, nextL)
			}
		}
	}

	if r.fwc != nil {
		for _, v2 := range r.fwc.psubs {
			logger.Infof("topic:%s level:%d", strings.Join(v2.tokens, "."), l)
		}
		if r.fwc.next != nil {
			s.printInfo("", tokens, r.fwc.next, nextL)
		}
	}

	if r.pwc != nil {
		for _, v2 := range r.pwc.psubs {
			logger.Infof("topic:%s level:%d", strings.Join(v2.tokens, "."), l)
		}
		if r.pwc.next != nil {
			s.printInfo("", tokens, r.pwc.next, nextL)
		}
	}
}

type lnt struct {
	l *level
	n *node
	t string
}

func (s *Match) remove(sub *subscription) error {
	var sfwc bool
	var n *node
	l := s.root

	var lnts [32]lnt
	levels := lnts[:0]

	for _, t := range sub.tokens {
		lt := len(t)
		if lt == 0 || sfwc {
			return ErrInvalidSubject
		}
		if l == nil {
			return ErrNotFound
		}
		if lt > 1 {
			n = l.nodes[t]
		} else {
			switch t[0] {
			case pwc:
				n = l.pwc
			case fwc:
				n = l.fwc
				sfwc = true
			default:
				n = l.nodes[t]
			}
		}
		if n != nil {
			levels = append(levels, lnt{l, n, t})
			l = n.next
		} else {
			l = nil
		}
	}
	removed, _ := s.removeFromNode(n, sub)
	if !removed {
		return ErrNotFound
	}

	s.count--
	s.removes++

	for i := len(levels) - 1; i >= 0; i-- {
		l, n, t := levels[i].l, levels[i].n, levels[i].t
		if n.isEmpty() {
			l.pruneNode(n, t)
		}
	}

	s.printInfo("after remove", sub.tokens, s.root, 1)
	return nil
}

func (s *Match) removeFromNode(n *node, sub *subscription) (found, last bool) {
	if n == nil {
		return false, true
	}
	_, found = n.psubs[sub]
	delete(n.psubs, sub)
	return found, len(n.psubs) == 0
}
