package etcd_locks

import (
	"context"
	etcd_storage "github.com/storage-lock/go-etcd-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var endpointsStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[string, *clientv3.Client] = storage_lock_factory.NewStorageLockFactoryBeanFactory[string, *clientv3.Client]()

func NewEtcdLockByEndpoints(ctx context.Context, endpoints []string, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetEtcdLockFactoryByEndpoints(ctx, endpoints)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewEtcdLockByEndpointsWithOptions(ctx context.Context, endpoints []string, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetEtcdLockFactoryByEndpoints(ctx, endpoints)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

// endpointsKey 将端点列表转为缓存 key
func endpointsKey(endpoints []string) string {
	key := ""
	for _, ep := range endpoints {
		key += ep + ","
	}
	return key
}

func GetEtcdLockFactoryByEndpoints(ctx context.Context, endpoints []string) (*storage_lock_factory.StorageLockFactory[*clientv3.Client], error) {
	key := endpointsKey(endpoints)
	return endpointsStorageLockFactoryBeanFactory.GetOrInit(ctx, key, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*clientv3.Client], error) {
		connectionManager := etcd_storage.NewEtcdConnectionManagerFromEndpoints(endpoints)
		options := etcd_storage.NewEtcdStorageOptions().SetConnectionManager(connectionManager)
		storage, err := newEtcdStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
