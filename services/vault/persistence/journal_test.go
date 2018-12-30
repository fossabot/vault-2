package persistence

import (
	"io/ioutil"
	"math"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jancajthaml-openbank/vault/model"

	money "gopkg.in/inf.v0"
)

func TestSnapshot_Update(t *testing.T) {
	name := "account_2"
	currency := "XRP"
	isBalanceCheck := false

	snapshotInitial := CreateAccount("/tmp/tenant", name, currency, isBalanceCheck)
	require.NotNil(t, snapshotInitial)

	loadedInitial := LoadAccount("/tmp/tenant", name)
	require.NotNil(t, loadedInitial)

	t.Log("Initial matches loaded")
	{
		assert.Equal(t, snapshotInitial.Balance, loadedInitial.Balance)
		assert.Equal(t, snapshotInitial.Promised, loadedInitial.Promised)
		assert.Equal(t, snapshotInitial.PromiseBuffer, loadedInitial.PromiseBuffer)
		assert.Equal(t, snapshotInitial.Version, loadedInitial.Version)
	}

	snapshotVersion1 := UpdateAccount("/tmp/tenant", name, snapshotInitial)
	require.NotNil(t, snapshotVersion1)

	loadedVersion1 := LoadAccount("/tmp/tenant", name)
	require.NotNil(t, loadedVersion1)

	t.Log("Updated matches loaded")
	{
		assert.Equal(t, snapshotVersion1.Balance, loadedVersion1.Balance)
		assert.Equal(t, snapshotVersion1.Promised, loadedVersion1.Promised)
		assert.Equal(t, snapshotVersion1.PromiseBuffer, loadedVersion1.PromiseBuffer)
		assert.Equal(t, snapshotVersion1.Version, loadedVersion1.Version)
	}

	t.Log("Updated is increment of version of initial by 1")
	{
		assert.Equal(t, loadedInitial.Version+1, loadedVersion1.Version)
		assert.Equal(t, snapshotInitial.Version+1, snapshotVersion1.Version)
	}
}

func TestSnapshot_RefuseOverflow(t *testing.T) {
	name := "xxx"
	currency := "XxX"
	isBalanceCheck := true

	snapshotLast := &model.Account{
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		PromiseBuffer:  model.NewTransactionSet(),
		Version:        int(math.MaxInt32),
		AccountName:    name,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	}

	snapshotNext := UpdateAccount("/tmp/tenant", name, snapshotLast)

	assert.Equal(t, snapshotLast.Version, snapshotNext.Version)
}

func TestSnapshot_PromiseBuffer(t *testing.T) {
	name := "yyy"
	currency := "yYy"
	isBalanceCheck := false

	expectedPromises := []string{"A", "B", "C", "D"}

	snapshot := &model.Account{
		Balance:        new(money.Dec),
		Promised:       new(money.Dec),
		PromiseBuffer:  model.NewTransactionSet(),
		Version:        0,
		AccountName:    name,
		Currency:       currency,
		IsBalanceCheck: isBalanceCheck,
	}

	snapshot.PromiseBuffer.Add(expectedPromises...)

	PersistAccount("/tmp/tenant", name, snapshot)

	loaded := LoadAccount("/tmp/tenant", name)

	if loaded == nil {
		t.Errorf("Expected to load snapshot got nil instead")
		return
	}

	assert.Equal(t, snapshot.Balance, loaded.Balance)
	assert.Equal(t, snapshot.Promised, loaded.Promised)
	assert.Equal(t, snapshot.Version, loaded.Version)
	assert.Equal(t, snapshot.AccountName, loaded.AccountName)
	assert.Equal(t, snapshot.Currency, loaded.Currency)
	assert.Equal(t, snapshot.IsBalanceCheck, loaded.IsBalanceCheck)

	require.Equal(t, len(expectedPromises), loaded.PromiseBuffer.Size())
	for _, v := range expectedPromises {
		assert.True(t, snapshot.PromiseBuffer.Contains(v))
	}
}

func BenchmarkAccountLoad(b *testing.B) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		b.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpDir)

	account := CreateAccount(tmpDir+"/tenant", "bench", "BNC", false)
	require.NotNil(b, account)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		LoadAccount(tmpDir+"/tenant", "bench")
	}
}