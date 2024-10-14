package hedis

import (
	"fmt"
	"os"
	"time"
)

func Persist() {

	for {
		time.Sleep(5 * time.Second)
		f, err := os.Create("dump.db")
		if err != nil {
			fmt.Println(err)
			continue
		}
		for k, v := range data {
			line := fmt.Sprintf("%s:%s", k, v)
			fmt.Fprintln(f, line)
		}
		err = f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
}
