package daily

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/pagefaultgames/pokerogue-server/db"
)

const secondsPerDay = 60 * 60 * 24

var (
	scheduler = gocron.NewScheduler(time.UTC)
	secret    []byte
)

func Init() error {
	var err error

	secret, err = os.ReadFile("secret.key")
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read daily seed secret: %s", err)
		}

		newSecret := make([]byte, 32)
		_, err := rand.Read(newSecret)
		if err != nil {
			return fmt.Errorf("failed to generate daily seed secret: %s", err)
		}

		err = os.WriteFile("secret.key", newSecret, 0400)
		if err != nil {
			return fmt.Errorf("failed to write daily seed secret: %s", err)
		}

		secret = newSecret
	}

	err = recordNewDaily()
	if err != nil {
		log.Print(err)
	}

	log.Printf("Daily Run Seed: %s", Seed())

	scheduler.Every(1).Day().At("00:00").Do(recordNewDaily())
	scheduler.StartAsync()

	return nil
}

func Seed() string {
	return base64.StdEncoding.EncodeToString(deriveSeed(time.Now().UTC()))
}

func deriveSeed(seedTime time.Time) []byte {
	day := make([]byte, 8)
	binary.BigEndian.PutUint64(day, uint64(seedTime.Unix()/secondsPerDay))

	hashedSeed := md5.Sum(append(day, secret...))

	return hashedSeed[:]
}

func recordNewDaily() error {
	err := db.TryAddDailyRun(Seed())
	if err != nil {
		return err
	}

	return nil
}