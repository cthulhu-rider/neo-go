package payload

import (
	"errors"

	"github.com/nspcc-dev/neo-go/pkg/config/netmode"
	"github.com/nspcc-dev/neo-go/pkg/core/block"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/util"
)

// MerkleBlock represents a merkle block packet payload.
type MerkleBlock struct {
	*block.Base
	Network netmode.Magic
	TxCount int
	Hashes  []util.Uint256
	Flags   []byte
}

// DecodeBinary implements Serializable interface.
func (m *MerkleBlock) DecodeBinary(br *io.BinReader) {
	m.Base = &block.Base{Network: m.Network}
	m.Base.DecodeBinary(br)

	txCount := int(br.ReadVarUint())
	if txCount > block.MaxContentsPerBlock {
		br.Err = block.ErrMaxContentsPerBlock
		return
	}
	m.TxCount = txCount
	br.ReadArray(&m.Hashes, m.TxCount)
	if txCount != len(m.Hashes) {
		br.Err = errors.New("invalid tx count")
	}
	m.Flags = br.ReadVarBytes((txCount + 7) / 8)
}

// EncodeBinary implements Serializable interface.
func (m *MerkleBlock) EncodeBinary(bw *io.BinWriter) {
	m.Base.EncodeBinary(bw)

	bw.WriteVarUint(uint64(m.TxCount))
	bw.WriteArray(m.Hashes)
	bw.WriteVarBytes(m.Flags)
}
