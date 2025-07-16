package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Restaurant struct {
	ID       string `json:"restaurant_id"`
	Domain   string `json:"iiko_web_domain"`
	Login    string `json:"iiko_web_login"`
	Password string `json:"iiko_web_password"`
}

type Output struct {
	RestaurantsWithIikoWebData []Restaurant `json:"restaurants_with_iikoweb_data"`
}

type ExistingFields struct {
	Domain   string `bson:"iiko_cloud.iiko_web_domain"`
	Login    string `bson:"iiko_cloud.iiko_web_login"`
	Password string `bson:"iiko_cloud.iiko_web_password"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите путь к JSON-файлу: ")
	filePath, _ := reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось открыть файл: %v\n", err)
		return
	}
	defer file.Close()

	var data Output
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при чтении JSON: %v\n", err)
		return
	}

	fmt.Print("Введите MongoDB URI: ")
	uri, _ := reader.ReadString('\n')
	uri = strings.TrimSpace(uri)

	fmt.Print("Введите имя базы данных: ")
	dbName, _ := reader.ReadString('\n')
	dbName = strings.TrimSpace(dbName)

	fmt.Print("Введите имя коллекции: ")
	collName, _ := reader.ReadString('\n')
	collName = strings.TrimSpace(collName)

	ctxConn, cancelConn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelConn()
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctxConn, clientOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения: %v\n", err)
		return
	}
	if err := client.Ping(ctxConn, readpref.Primary()); err != nil {
		fmt.Fprintf(os.Stderr, "Ping не прошёл: %v\n", err)
		return
	}
	fmt.Println("✅ Успешное подключение к MongoDB")

	collection := client.Database(dbName).Collection(collName)

	var total, skipped, updated int

	for _, r := range data.RestaurantsWithIikoWebData {
		total++
		oid, err := primitive.ObjectIDFromHex(r.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Некорректный id: \"%s\", сообщение: \"%v\" пропускаем\n", r.ID, err)
			skipped++
			continue
		}
		filter := bson.M{"_id": oid}

		var exist ExistingFields
		err = collection.FindOne(
			context.Background(), filter,
			options.FindOne().SetProjection(bson.M{
				"iiko_web_domain":   1,
				"iiko_web_login":    1,
				"iiko_web_password": 1,
				"_id":               0,
			}),
		).Decode(&exist)
		if err == mongo.ErrNoDocuments {
			fmt.Printf("id: \"%s\" не найден в базе, пропускаем\n", r.ID)
			skipped++
			continue
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка FindOne id=%s: %v\n", r.ID, err)
			skipped++
			continue
		}

		if exist.Domain == r.Domain && exist.Login == r.Login && exist.Password == r.Password {
			fmt.Printf("id: \"%s\", данные совпадают, пропускаем\n", r.ID)
			skipped++
			continue
		}

		update := bson.M{"$set": bson.M{
			"iiko_cloud.iiko_web_domain":   r.Domain,
			"iiko_cloud.iiko_web_login":    r.Login,
			"iiko_cloud.iiko_web_password": r.Password,
		}}
		ctxUpd, cancelUpd := context.WithTimeout(context.Background(), 15*time.Second)
		_, err = collection.UpdateOne(ctxUpd, filter, update)
		cancelUpd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка UpdateOne _id=%s: %v\n", r.ID, err)
			skipped++
			continue
		}
		fmt.Printf("id: %s, обновлено\n", r.ID)
		updated++
	}

	fmt.Printf("\nВсего записей: %d, обновлено: %d, пропущено: %d\n", total, updated, skipped)
	fmt.Println("Завершено.")
}
