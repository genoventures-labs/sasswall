package tarpit

import (
	"math/rand"
	"time"
)

const MaxTotalDelay = 8 * time.Second

type Config struct {
	DelayMin    time.Duration
	DelayMax    time.Duration
	HoneyBoost  time.Duration
	StrikeBoost time.Duration
}

type Calculator struct {
	cfg Config
	rng *rand.Rand
}

func New(cfg Config) *Calculator {
	return &Calculator{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *Calculator) Delay(sequenceScore int, isHoney bool, strikes int) time.Duration {
	base := c.randomBetween(c.cfg.DelayMin, c.cfg.DelayMax)
	if isHoney {
		base += c.cfg.HoneyBoost
	}
	if strikes > 0 {
		base += time.Duration(strikes) * c.cfg.StrikeBoost
	}
	multiplier := 1.0
	switch {
	case sequenceScore >= 8:
		multiplier = 2.2 // +120%
	case sequenceScore >= 4:
		multiplier = 1.4 // +40%
	}
	total := time.Duration(float64(base) * multiplier)
	if total > MaxTotalDelay {
		return MaxTotalDelay
	}
	return total
}

func (c *Calculator) randomBetween(min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	d := max - min
	return min + time.Duration(c.rng.Int63n(int64(d)+1))
}
