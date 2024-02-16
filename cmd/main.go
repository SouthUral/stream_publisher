package main

import (
	"context"
	"fmt"

	"streamPublisher/internal/amqp"
	"streamPublisher/internal/core"
	"streamPublisher/internal/psql"
	"time"
)

func main() {
	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", "kovalenko", "kovalenko", "localhost", "5433", "db_test")
	rbUrl := fmt.Sprintf("rabbitmq-stream://%s:%s@%s:%s/%s", "test_user", "rmpassword", "localhost", "5552", "asd")

	pgObj := psql.InitPsql(pgUrl, 10, 5, 20, 2)
	amqpObj := amqp.InitRabbitWorker(rbUrl, "test_produser", "messages_stream")

	coreObj := core.InitCore(2000, pgObj.GetCh(), amqpObj.GetCh())

	time.Sleep(2 * time.Second)
	coreObj.StartProcess()

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Minute)
	<-ctx.Done()
	err := fmt.Errorf("тест закончен")
	pgObj.Shutdown(err)
	coreObj.Shutdown(err)
}
