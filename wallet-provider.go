package wallet_provider

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	evmWalletService "github.com/zd4r/wallet-provider/service/evm_wallet"
	"github.com/zd4r/wallet-provider/storage/sqlite"
	evmWalletStore "github.com/zd4r/wallet-provider/store/evm_wallet"
	"github.com/zd4r/wallet-provider/store/passphrase"
	"golang.org/x/term"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultKeystoreDirPath = "/.ethereum/keystore"
	defaultDatabasePath    = "/.ethereum/wallet-manager.db"
)

func New(keystoreDirPath, databasePath string) (*evmWalletService.Service, error) {
	if keystoreDirPath == "" {
		keystoreDirPath = defaultKeystoreDirPath
	}

	if databasePath == "" {
		databasePath = defaultDatabasePath
	}

	cxt := context.Background()

	// init global passphrase store
	pp := passphrase.New()

	// set passphrase
	password, err := readPassphrase()
	if err != nil {
		return nil, fmt.Errorf("failed to read passphrase: %w", err)
	}
	pp.Set(password)

	// get $HOME
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get $HOME: %w", err)
	}

	// init keystore
	ks := keystore.NewKeyStore(
		fmt.Sprintf("%s%s", homeDir, keystoreDirPath), // TODO: change path to $HOME/.ethereum/keystore
		keystore.StandardScryptN,
		keystore.StandardScryptP,
	)

	// init data storage
	storage, err := sqlite.NewWithContext(cxt, fmt.Sprintf("%s%s", homeDir, databasePath))
	if err != nil {
		return nil, fmt.Errorf("failed to init data storage: %w", err)
	}

	// init wallet service
	evmWalletSrv := evmWalletService.New(
		evmWalletStore.New(storage),
		ks,
		pp,
	)
	if err := evmWalletSrv.CheckAccess(); err != nil {
		return nil, fmt.Errorf("failed to init wallet service: %w", err)
	}

	return evmWalletSrv, nil
}

func readPassphrase() ([]byte, error) {
	defer fmt.Printf("\n\n")

	fmt.Print("password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to term.ReadPassword: %w", err)
	}

	return password, nil
}
