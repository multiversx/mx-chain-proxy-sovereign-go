package resultsParser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data/transaction"

	"github.com/stretchr/testify/require"
)

var testPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

func TestResultsParser_ParseUntypedOutcome(t *testing.T) {
	t.Parallel()

	t.Run("should parse contract outcome, on easily found result with return data", func(t *testing.T) {
		t.Parallel()

		transactionResult := &transaction.ApiTransactionResult{
			SmartContractResults: []*transaction.ApiSmartContractResult{
				{
					Nonce:         42,
					Data:          "@6f6b@03",
					ReturnMessage: "foobar",
				},
			},
		}

		outcome, err := ParseResultOutcome(transactionResult, testPubkeyConverter)
		require.NoError(t, err)
		require.Equal(t, Ok, outcome.ReturnCode)
		require.Equal(t, "foobar", outcome.ReturnMessage)
		require.Equal(t, outcome.Values, []*bytes.Buffer{bytes.NewBuffer([]byte("03"))})
	})

	t.Run("should parse contract outcome, on signal error", func(t *testing.T) {
		t.Parallel()

		transactionResult := &transaction.ApiTransactionResult{
			Logs: &transaction.ApiLogs{
				Address: "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				Events: []*transaction.Events{
					{
						Identifier: "signalError",
						Topics: [][]byte{
							[]byte("something happened"),
						},
						Data: []byte("@75736572206572726f72@07"),
					},
				},
			},
		}

		outcome, err := ParseResultOutcome(transactionResult, testPubkeyConverter)
		require.NoError(t, err)
		require.Equal(t, UserError, outcome.ReturnCode)
		require.Equal(t, outcome.Values, []*bytes.Buffer{bytes.NewBuffer([]byte("07"))})
	})

	t.Run("should parse contract outcome, on too much gas warning", func(t *testing.T) {
		t.Parallel()

		transactionResult := &transaction.ApiTransactionResult{
			Logs: &transaction.ApiLogs{
				Address: "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				Events: []*transaction.Events{
					{
						Identifier: "writeLog",
						Topics: [][]byte{
							[]byte("@too much gas provided for processing: gas provided = 596384500, gas used = 733010"),
						},
						Data: []byte("@6f6b"),
					},
				},
			},
		}

		outcome, err := ParseResultOutcome(transactionResult, testPubkeyConverter)
		require.NoError(t, err)
		require.Equal(t, Ok, outcome.ReturnCode)
		require.Equal(t, "@too much gas provided for processing: gas provided = 596384500, gas used = 733010", outcome.ReturnMessage)
		require.Empty(t, outcome.Values)
	})

	t.Run("should parse contract outcome, on write log where first topic equals address", func(t *testing.T) {
		t.Parallel()

		transactionResult := &transaction.ApiTransactionResult{
			Sender: "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
			Logs: &transaction.ApiLogs{
				Events: []*transaction.Events{
					{
						Identifier: "writeLog",
						Topics: [][]byte{
							[]byte("ZXJkMXF5dTV3dGhsZHpyOHd4NWM5dWNnOGtqYWdnMGpmczUzczhucjN6cHozaHlwZWZzZGQ4c3N5Y3I2dGg="),
						},
						Data: []byte("@6f6b="),
					},
				},
			},
		}

		outcome, err := ParseResultOutcome(transactionResult, testPubkeyConverter)
		require.NoError(t, err)
		require.Equal(t, Ok, outcome.ReturnCode)
		require.Empty(t, outcome.Values)
	})
}

// Tested on 1st July 2024 with 10k transactions.
func TestResultsParser_RealWorld(t *testing.T) {
	//t.Skip()

	filePath := "./transactions.json"

	txs, err := readJSONFromFile(filePath)
	if err != nil {
		fmt.Printf("Error reading from file: %v\n", err)
		return
	}

	var nilOutcomes []*transaction.ApiTransactionResult
	for i, tx := range txs {
		outcome, err := ParseResultOutcome(tx, testPubkeyConverter)
		if err != nil {
			panic(fmt.Errorf("error parsing results: %v %d\n", err, i))
		}

		if outcome == nil {
			nilOutcomes = append(nilOutcomes, tx)
		}
	}
}

func readJSONFromFile(filePath string) ([]*transaction.ApiTransactionResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var txs []*transaction.ApiTransactionResult
	if err := json.Unmarshal(byteValue, &txs); err != nil {
		return nil, err
	}

	return txs, nil
}
