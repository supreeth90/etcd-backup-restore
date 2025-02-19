package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdResponse struct {
	Header struct {
		Revision int64 `json:"revision"`
	} `json:"header"`
	Kvs []struct {
		Key            string `json:"key"`
		CreateRevision int64  `json:"create_revision"`
		ModRevision    int64  `json:"mod_revision"`
		Value          string `json:"value"`
	} `json:"kvs"`
}

func GetEtcdClient() (*clientv3.Client, error) {
	// Create etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create etcd client: %v", err)
	}

	return client, err
}

func Backup(client *clientv3.Client, dumpFile string) {
	// Get all keys
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a map to store key-value pairs
	kvMap := make(map[string]string)

	resp, err := client.Get(ctx, "", clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Failed to get keys: %v", err)
	}

	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		value := base64.StdEncoding.EncodeToString(kv.Value)
		kvMap[key] = value
	}

	jsonData, err := json.MarshalIndent(kvMap, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	//resp, err := client.Get(ctx, "\x00", clientv3.WithRange("\x00"))
	//if err != nil {
	//	log.Fatalf("Failed to get keys: %v", err)
	//}
	//
	//
	//
	//// Iterate through the key-value pairs
	//for _, kv := range resp.Kvs {
	//	kvMap[string(kv.Key)] = string(kv.Value)
	//}
	//
	//// Convert map to JSON
	//jsonData, err := json.MarshalIndent(kvMap, "", "  ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//}

	// Write JSON to file
	err = os.WriteFile(dumpFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("Failed to write JSON file: %v", err)
	}

	fmt.Println("Successfully wrote etcd data to etcd_data.json")
}

func Restore(client *clientv3.Client, dataMap map[string]string) {
	log.Println("In restore")
	//defer client.Close()

	// Create a map of key/value pairs
	//dataMap := map[string]string{
	//	"/users/12345/email": "new.email@example.com",
	//"/users/12345/phone": "987-654-3210",
	//"/users/67890/name":  "John Doe",
	//}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("In restore, set context done")

	// Prepare transaction operations
	var cmps []clientv3.Cmp
	var puts []clientv3.Op
	var gets []clientv3.Op
	log.Println("In restore, set transaction operations done")

	for key, value := range dataMap {
		cmps = append(cmps, clientv3.Compare(clientv3.ModRevision(key), "=", 0))
		puts = append(puts, clientv3.OpPut(key, value))
		gets = append(gets, clientv3.OpGet(key))

		log.Println("In restore, before performing transactional put")
		// Perform transactional put

		//tx := client.Txn(ctx)
		txnResp, err := client.Txn(ctx).If(cmps...).Then(puts...).Else(gets...).Commit()
		//txnResp, err := client.Txn(ctx).Then(puts...).Else(gets...).Commit()
		log.Println("In restore, after performing transactional put")

		if err != nil {
			log.Fatalf("Transaction failed: %v", err)
		}

		log.Println("In restore, no error in performing transactional put")
		if txnResp.Succeeded {
			fmt.Println("Transaction succeeded. Keys added:")
			for key := range dataMap {
				fmt.Printf("- %s\n", key)
			}
		} else {
			fmt.Println("Transaction failed. Existing keys:")
			for _, ev := range txnResp.Responses {
				for _, kv := range ev.GetResponseRange().Kvs {
					fmt.Printf("%s: %s\n", string(kv.Key), string(kv.Value))
				}
			}
		}
	}

	log.Println("In restore,  performing transactional done")
}

func parseEtcdJSON(jsonStr string) (map[string]string, error) {
	var response EtcdResponse
	err := json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	result := make(map[string]string)
	for _, kv := range response.Kvs {
		decodedKey, err := base64.StdEncoding.DecodeString(kv.Key)
		if err != nil {
			return nil, fmt.Errorf("error decoding key: %v", err)
		}

		decodedValue, err := base64.StdEncoding.DecodeString(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("error decoding value: %v", err)
		}

		result[string(decodedKey)] = string(decodedValue)
		//result[kv.Key] = kv.Value
	}

	return result, nil
}
