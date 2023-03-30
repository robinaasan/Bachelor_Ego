package verifyreport

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"github.com/edgelesssys/ego/enclave"
)

//Containts functions used by both the orderingservice and runtime in localattestation
//cert_pubKey -> pubKey for ordering_service
//cert_pubKey -> cert for runtime

func VerifyReport(reportBytes []byte, cert_pubKey []byte) error {
	report, err := enclave.VerifyLocalReport(reportBytes)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(cert_pubKey)
	if !bytes.Equal(report.Data[:len(hash)], hash[:]) {
		return errors.New("report data doesn't match the server certificate's hash")
	}

	// We expect the other enclave to be signed with the same key.

	selfReport, err := enclave.GetSelfReport()
	if err != nil {
		return err
	}

	if !bytes.Equal(report.SignerID, selfReport.SignerID) {
		return errors.New("invalid signer")
	}
	if binary.LittleEndian.Uint16(report.ProductID) != 2 {
		return errors.New("invalid product")
	}
	if report.SecurityVersion < 1 {
		return errors.New("invalid security version")
	}
	if report.Debug && !selfReport.Debug {
		return errors.New("other party is a debug enclave")
	}

	return nil
}
