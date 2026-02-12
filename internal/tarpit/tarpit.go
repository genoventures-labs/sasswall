package tarpit

import (
	"crypto/sha256"
	"encoding/binary"
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
	return c.DelayWithPresence(sequenceScore, isHoney, strikes, 1.0, "", false, 0)
}

func (c *Calculator) DelayWithPresence(sequenceScore int, isHoney bool, strikes int, presenceMultiplier float64, signatureID string, timingSignature bool, jitterBandMS int) time.Duration {
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
	if presenceMultiplier > 0 {
		multiplier *= presenceMultiplier
	}
	total := time.Duration(float64(base) * multiplier)
	if timingSignature && jitterBandMS > 0 && signatureID != "" {
		total += c.signatureJitter(signatureID, jitterBandMS)
	}
	if total > MaxTotalDelay {
		return MaxTotalDelay
	}
	if total < 0 {
		return 0
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

func (c *Calculator) signatureJitter(signatureID string, bandMS int) time.Duration {
	h := sha256.Sum256([]byte(signatureID))
	v := binary.BigEndian.Uint32(h[:4])
	span := int(v % uint32((bandMS*2)+1))
	jitterMS := span - bandMS
	return time.Duration(jitterMS) * time.Millisecond
}
