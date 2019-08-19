package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/zshi-redhat/kube-ptp-daemon/logging"
)

const (
	_ETHTOOL_HARDWARE_RECEIVE_CAP = "hardware-receive"
	_ETHTOOL_HARDWARE_TRANSMIT_CAP = "hardware-transmit"
	_ETHTOOL_HARDWARE_RAW_CLOCK_CAP = "hardware-raw-clock"
	_ETHTOOL_RX_HARDWARE_FLAG  = "SOF_TIMESTAMPING_RX_HARDWARE"
	_ETHTOOL_TX_HARDWARE_FLAG  = "SOF_TIMESTAMPING_TX_HARDWARE"
	_ETHTOOL_RAW_HARDWARE_FLAG = "SOF_TIMESTAMPING_RAW_HARDWARE"
)

func ethtoolInstalled() bool {
	_, err := exec.LookPath("ethtool")
	return err == nil
}

func netParseEthtoolTimeStampFeature(cmdOut *bytes.Buffer) bool {
	var hardRxEnabled bool
	var hardTxEnabled bool
	var hardRawEnabled bool

	logging.Debugf("cmd output for %v", cmdOut)
	scanner := bufio.NewScanner(cmdOut)
	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), "\t")
		parts := strings.Fields(line)
		if parts[0] == _ETHTOOL_HARDWARE_RECEIVE_CAP {
			hardRxEnabled = parts[1] == _ETHTOOL_RX_HARDWARE_FLAG
		}
		if parts[0] == _ETHTOOL_HARDWARE_TRANSMIT_CAP {
			hardTxEnabled = parts[1] == _ETHTOOL_TX_HARDWARE_FLAG
		}
		if parts[0] == _ETHTOOL_HARDWARE_RAW_CLOCK_CAP {
			hardRawEnabled = parts[1] == _ETHTOOL_RAW_HARDWARE_FLAG
		}
	}
	return hardRxEnabled && hardTxEnabled && hardRawEnabled
}

func DiscoverPTPDevices() ([]string, error) {
	var out bytes.Buffer
	nics := make([]string, 0)

	if !ethtoolInstalled() {
                return nics, fmt.Errorf("discoverDevices(): ethtool not installed. Cannot grab NIC capabilities")
	}

	ethtoolPath, _ := exec.LookPath("ethtool")

	net, err := ghw.Network()
        if err != nil {
                return nics, fmt.Errorf("discoverDevices(): error getting network info: %v", err)
        }

        for _, dev := range net.NICs {
		logging.Debugf("grabbing NIC timestamp capability for %v", dev.Name)
		cmd := exec.Command(ethtoolPath, "-T", dev)
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			logging.Debugf("could not grab NIC timestamp capability for %v: %v", dev, err)
		}
		if netParseEthtoolTimeStampFeature(&out) {
			nics = append(nics, device)
		}
	}
	return nics, nil
}
