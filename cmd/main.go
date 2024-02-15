package main

import (
	"context"
	"fmt"

	"streamPublisher/internal/core"
	"streamPublisher/internal/psql"
	"time"
)

func main() {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", "kovalenko", "kovalenko", "localhost", "5433", "db_test")
	coreObj := core.InitCore(2000)
	pgObj := psql.InitPsql(pgUrl, coreObj.GetStorageCh(), 10, 5, 20, 2)

	time.Sleep(2 * time.Second)
	coreObj.StartProcess()

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)
	<-ctx.Done()
	err := fmt.Errorf("тест закончен")
	pgObj.Shutdown(err)
	coreObj.Shutdown(err)
}
