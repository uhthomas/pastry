package pastry_test

import (
	"testing"
)

type testerror string

func (err testerror) Error() string { return string(err) }

type ErrorReader struct{ err error }

func (err ErrorReader) Read([]byte) (int, error) { return 0, err.err }

type ErrorWriter struct{ err error }

func (err ErrorWriter) Write([]byte) (int, error) { return 0, err.err }

func TestNode_Handshake(t *testing.T) {
	t.Skip("needs to be refactored")
	/*	t.Run("should return when generating key-pair", func(t *testing.T) {
			n, err := pastry.New()
			require.NoError(t, err)

			const someError testerror = "some-error"

			_, _, err = n.Handshake(nil, nil, ErrorReader{err: someError})
			require.Error(t, err)
			assert.Equal(t, someError, err)
		})

		t.Run("should return error when writing to w", func(t *testing.T) {
			n, err := pastry.New()
			require.NoError(t, err)

			const someError testerror = "some-error"

			_, _, err = n.Handshake(
				ErrorWriter{err: someError},
				nil,
				rand.Reader,
			)
			require.Error(t, err)
			assert.Equal(t, someError, err)
		})

		t.Run("should return error when reading from r", func(t *testing.T) {
			n, err := pastry.New()
			require.NoError(t, err)

			const someError testerror = "some-error"

			_, _, err = n.Handshake(
				ioutil.Discard,
				ErrorReader{err: someError},
				rand.Reader,
			)
			require.Error(t, err)
			assert.Equal(t, someError, err)
		})*/
}
