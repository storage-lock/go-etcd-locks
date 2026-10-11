package etcd_locks

import (
	"context"
	etcd_storage "github.com/storage-lock/go-etcd-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var clientStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[*clientv3.Client, *clientv3.Client] = storage_lock_factory.NewStorageLockFactoryBeanFactory[*clientv3.Client, *clientv3.Client]()

// newEtcdStorage 指向 etcd_storage.NewEtcdStorage，包级变量以便单测注入构造失败，
// 覆盖 GetOrInit initFunc 的错误分支（默认实现在 options 合法、本机服务正常时永不失败）。
var newEtcdStorage = etcd_storage.NewEtcdStorage

func NewEtcdLockByClient(ctx context.Context, client *clientv3.Client, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetEtcdLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewEtcdLockByClientWithOptions(ctx context.Context, client *clientv3.Client, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetEtcdLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

func GetEtcdLockFactoryByClient(ctx context.Context, client *clientv3.Client) (*storage_lock_factory.StorageLockFactory[*clientv3.Client], error) {
	return clientStorageLockFactoryBeanFactory.GetOrInit(ctx, client, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*clientv3.Client], error) {
		connectionManager := etcd_storage.NewEtcdConnectionManagerFromClient(client)
		options := etcd_storage.NewEtcdStorageOptions().SetConnectionManager(connectionManager)
		storage, err := newEtcdStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
