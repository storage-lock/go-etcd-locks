package etcd_locks

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewEtcdLockByEndpoints(t *testing.T) {
	envName := "STORAGE_LOCK_ETCD_ENDPOINTS"
	endpoints := os.Getenv(envName)
	if endpoints == "" {
		t.Skip("STORAGE_LOCK_ETCD_ENDPOINTS not set, skipping etcd test")
	}

	ctx := context.Background()
	lockId := "test-etcd-lock"

	factory, err := GetEtcdLockFactoryByEndpoints(ctx, []string{endpoints})
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	lock, err := NewEtcdLockByEndpoints(ctx, []string{endpoints}, lockId)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	// 加锁
	ownerId := "test-owner"
	err = lock.Lock(ctx, ownerId)
	assert.Nil(t, err)

	// 解锁
	err = lock.UnLock(ctx, ownerId)
	assert.Nil(t, err)
}
