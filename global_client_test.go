package etcd_locks

import (
	"context"
	"errors"
	"testing"
	"time"

	etcd_storage "github.com/storage-lock/go-etcd-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testEtcdClient(t *testing.T) *clientv3.Client {
	t.Helper()
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	assert.Nil(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 探活：确认本机 etcd 真可用，否则跳过成功路径
	if _, err := client.Status(ctx, "http://127.0.0.1:2379"); err != nil {
		client.Close()
		t.Skipf("本机 etcd 不可用，跳过: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestNewEtcdLockByClient(t *testing.T) {
	client := testEtcdClient(t)

	ctx := context.Background()
	factory, err := GetEtcdLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetEtcdLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	lock, err := NewEtcdLockByClient(ctx, client, "test-etcd-lock-client")
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	assert.Nil(t, lock.Lock(ctx, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx, "owner-1"))
}

func TestNewEtcdLockByClientWithOptions(t *testing.T) {
	client := testEtcdClient(t)

	options := storage_lock.NewStorageLockOptionsWithLockId("test-etcd-lock-client-opt")
	lock, err := NewEtcdLockByClientWithOptions(context.Background(), client, options)
	assert.Nil(t, err)
	assert.NotNil(t, lock)
}

func TestEtcdLockFactory_Errors(t *testing.T) {
	old := newEtcdStorage
	newEtcdStorage = func(ctx context.Context, options *etcd_storage.EtcdStorageOptions) (*etcd_storage.EtcdStorage, error) {
		return nil, errors.New("boom")
	}
	defer func() { newEtcdStorage = old }()

	// ByClient 错误分支（用独立 client 避免缓存污染）
	boomClient, err := clientv3.New(clientv3.Config{Endpoints: []string{"http://127.0.0.1:2379"}})
	assert.Nil(t, err)
	defer boomClient.Close()
	_, err = GetEtcdLockFactoryByClient(context.Background(), boomClient)
	assert.NotNil(t, err)
	_, err = NewEtcdLockByClient(context.Background(), boomClient, "x")
	assert.NotNil(t, err)
	_, err = NewEtcdLockByClientWithOptions(context.Background(), boomClient, storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)

	// ByEndpoints 错误分支（用独立 endpoints 避免缓存污染）
	dead := []string{"http://127.0.0.1:19999-err"}
	_, err = GetEtcdLockFactoryByEndpoints(context.Background(), dead)
	assert.NotNil(t, err)
	_, err = NewEtcdLockByEndpoints(context.Background(), dead, "x")
	assert.NotNil(t, err)
	_, err = NewEtcdLockByEndpointsWithOptions(context.Background(), dead, storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)
}

func TestNewEtcdLockByEndpoints_Local(t *testing.T) {
	endpoints := []string{"http://127.0.0.1:2379"}
	probe, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 3 * time.Second})
	assert.Nil(t, err)
	defer probe.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := probe.Status(ctx, endpoints[0]); err != nil {
		t.Skipf("本机 etcd 不可用，跳过: %v", err)
	}

	ctx2 := context.Background()
	factory, err := GetEtcdLockFactoryByEndpoints(ctx2, endpoints)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetEtcdLockFactoryByEndpoints(ctx2, []string{"http://127.0.0.1:2379"})
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	lock, err := NewEtcdLockByEndpoints(ctx2, endpoints, "test-etcd-lock-endpoints")
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	options := storage_lock.NewStorageLockOptionsWithLockId("test-etcd-lock-endpoints-opt")
	lock2, err := NewEtcdLockByEndpointsWithOptions(ctx2, endpoints, options)
	assert.Nil(t, err)
	assert.NotNil(t, lock2)

	assert.Nil(t, lock.Lock(ctx2, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx2, "owner-1"))
}

func TestEndpointsKey(t *testing.T) {
	assert.Equal(t, "a,b,", endpointsKey([]string{"a", "b"}))
	assert.Equal(t, "", endpointsKey(nil))
}
