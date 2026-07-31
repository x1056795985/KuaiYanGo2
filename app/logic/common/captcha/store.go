package captcha

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"

	"server/app/global"
	"server/app/logic/common/cache"
)

const (
	cachePrefix = "captcha:"
	defaultTTL  = 5 * time.Minute
	lockCount   = 64
)

// Store adapts the application's cache to base64Captcha and serializes
// operations for the same challenge to prevent replays within one process.
type Store struct {
	cache cache.Cache
	locks [lockCount]sync.Mutex
}

func NewStore(backend cache.Cache) *Store {
	return &Store{cache: backend}
}

// VerificationCodes is the shared store used by image and SMS challenges.
// Its backend is resolved lazily because global.H缓存 is initialized in main.
var VerificationCodes = NewStore(nil)

func (s *Store) backend() cache.Cache {
	if s.cache != nil {
		return s.cache
	}
	return global.H缓存
}

func (s *Store) lock(id string) func() {
	var hash uint32 = 2166136261
	for i := 0; i < len(id); i++ {
		hash ^= uint32(id[i])
		hash *= 16777619
	}
	lock := &s.locks[hash%lockCount]
	lock.Lock()
	return lock.Unlock
}

// Set implements base64Captcha.Store.
func (s *Store) Set(id, value string) {
	s.set(id, value)
}

func (s *Store) set(id string, value any) {
	if id == "" {
		return
	}
	backend := s.backend()
	if backend == nil {
		return
	}
	unlock := s.lock(id)
	defer unlock()
	backend.Set(cachePrefix+id, value, defaultTTL)
}

// Get returns a string challenge. It never panics when a cache entry has an
// unexpected type. clear is useful for callers that intentionally consume on read.
func (s *Store) Get(id string, clear bool) string {
	unlock := s.lock(id)
	defer unlock()

	value, ok := s.getLocked(id)
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	if clear {
		s.deleteLocked(id)
	}
	return text
}

// Verify implements base64Captcha.Store. A successful code is always consumed;
// a failed attempt no longer deletes a still-valid code.
func (s *Store) Verify(id, answer string, _ bool) bool {
	if id == "" || answer == "" {
		return false
	}
	unlock := s.lock(id)
	defer unlock()

	value, ok := s.getLocked(id)
	if !ok {
		return false
	}
	expected, ok := value.(string)
	if !ok || expected != answer {
		return false
	}
	s.deleteLocked(id)
	return true
}

func (s *Store) getLocked(id string) (any, bool) {
	backend := s.backend()
	if backend == nil {
		return nil, false
	}
	return backend.Get(cachePrefix + id)
}

func (s *Store) deleteLocked(id string) {
	if backend := s.backend(); backend != nil {
		backend.Delete(cachePrefix + id)
	}
}

func randomToken(length int, alphabet string) (string, error) {
	if length <= 0 || len(alphabet) == 0 || len(alphabet) > 256 {
		return "", errors.New("invalid random token parameters")
	}
	result := make([]byte, length)
	random := make([]byte, length*2)
	limit := 256 - 256%len(alphabet)
	for i := 0; i < length; {
		if _, err := rand.Read(random); err != nil {
			return "", err
		}
		for _, value := range random {
			if int(value) >= limit {
				continue
			}
			result[i] = alphabet[int(value)%len(alphabet)]
			i++
			if i == length {
				break
			}
		}
	}
	return string(result), nil
}

func randomInt(max int) (int, error) {
	if max <= 0 || max > 256 {
		return 0, errors.New("random range must be between 1 and 256")
	}
	limit := 256 - 256%max
	var value [1]byte
	for {
		if _, err := rand.Read(value[:]); err != nil {
			return 0, err
		}
		if int(value[0]) < limit {
			return int(value[0]) % max, nil
		}
	}
}
