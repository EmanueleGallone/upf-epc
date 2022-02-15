// SPDX-License-Identifier: Apache-2.0
// Copyright 2022 Open Networking Foundation

package pfcpiface

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"testing"

	//nolint:staticcheck // Ignore SA1019.
	// Upgrading to google.golang.org/protobuf/proto is not a drop-in replacement,
	// as also P4Runtime stubs are based on the deprecated proto.
	"github.com/golang/protobuf/proto"
	p4ConfigV1 "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/stretchr/testify/require"
)

//nolint:unused
func setupNewTranslator(t *testing.T) *P4rtTranslator {
	p4infoBytes, err := ioutil.ReadFile("../conf/p4/bin/p4info.txt")
	require.NoError(t, err)

	var p4Config p4ConfigV1.P4Info

	err = proto.UnmarshalText(string(p4infoBytes), &p4Config)
	require.NoError(t, err)

	return newP4RtTranslator(p4Config)
}

func Test_actionID(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		translator *P4rtTranslator
		want       uint32
	}{
		{name: "get NoAction",
			args:       "NoAction",
			translator: setupNewTranslator(t),
			want:       uint32(21257015),
		},
		{name: "non existing action",
			args:       "qwerty",
			translator: setupNewTranslator(t),
			want:       uint32(0),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := tt.translator.actionID(tt.args)
				require.Equal(t, tt.want, got)
			},
		)
	}
}

func Test_tableID(t *testing.T) {
	tests := []struct {
		name       string
		args       string
		translator *P4rtTranslator
		want       uint32
	}{
		{name: "Existing table",
			args:       "PreQosPipe.Routing.routes_v4",
			translator: setupNewTranslator(t),
			want:       uint32(39015874),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := tt.translator.tableID(tt.args)
				require.Equal(t, tt.want, got)
			},
		)
	}
}

func Test_getCounterByName(t *testing.T) {
	type args struct {
		counterName string
		counterID   uint32
		translator  *P4rtTranslator
	}

	type want struct {
		counterName string
		counterID   uint32
	}

	tests := []struct {
		name    string
		args    *args
		want    *want
		wantErr bool
	}{
		{name: "Existing counter",
			args: &args{
				counterName: "PreQosPipe.pre_qos_counter",
				counterID:   uint32(315693181),
				translator:  setupNewTranslator(t),
			},
			want: &want{
				counterName: "PreQosPipe.pre_qos_counter",
				counterID:   uint32(315693181),
			},
			wantErr: false,
		},
		{name: "Existing counter",
			args: &args{
				counterName: "PostQosPipe.post_qos_counter",
				counterID:   uint32(302958180),
				translator:  setupNewTranslator(t),
			},
			want: &want{
				counterName: "PostQosPipe.post_qos_counter",
				counterID:   uint32(302958180),
			},
			wantErr: false,
		},
		{name: "Non existing counter",
			args: &args{
				counterName: "testttt",
				counterID:   uint32(0),
				translator:  setupNewTranslator(t),
			},
			want:    &want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.args.translator.getCounterByName(tt.want.counterName)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.Equal(t, tt.want.counterID, got.GetPreamble().GetId())
					require.Equal(t, tt.want.counterName, got.GetPreamble().Name)
				}
			},
		)
	}
}

func Test_getTableByID(t *testing.T) {
	type args struct {
		tableID    uint32
		tableName  string
		translator *P4rtTranslator
	}

	type want struct {
		tableID   uint32
		tableName string
	}

	tests := []struct {
		name    string
		args    *args
		want    *want
		wantErr bool
	}{
		{name: "Existing table",
			args: &args{
				tableID:    39015874,
				tableName:  "PreQosPipe.Routing.routes_v4",
				translator: setupNewTranslator(t),
			},
			want: &want{
				tableID:   39015874,
				tableName: "PreQosPipe.Routing.routes_v4",
			},
			wantErr: false,
		},
		{name: "non existing table",
			args: &args{
				tableID:    999999,
				translator: setupNewTranslator(t),
			},
			want:    &want{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.args.translator.getTableByID(tt.want.tableID)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.Equal(t, tt.want.tableID, got.GetPreamble().GetId())
					require.Equal(t, tt.want.tableName, got.GetPreamble().GetName())
				}
			},
		)
	}
}

func Test_getMatchFieldByName(t *testing.T) {
	ts := setupNewTranslator(t)

	mockTable, err := ts.getTableByID(47204971) //acls table
	require.NoError(t, err)

	type args struct {
		table          *p4ConfigV1.Table
		matchFieldName string
	}

	type want struct {
		name  string
		id    uint32
		match p4ConfigV1.MatchField_MatchType
	}

	tests := []struct {
		name    string
		args    *args
		want    *want
		wantErr bool
	}{
		{name: "Existing match Field",
			args: &args{
				table:          mockTable,
				matchFieldName: "inport",
			},
			want: &want{
				name:  "inport",
				id:    1,
				match: p4ConfigV1.MatchField_TERNARY,
			},
		},
		{name: "Existing match Field",
			args: &args{
				table:          mockTable,
				matchFieldName: "ipv4_dst",
			},
			want: &want{
				name:  "ipv4_dst",
				id:    7,
				match: p4ConfigV1.MatchField_TERNARY,
			},
		},
		{name: "non existing match field",
			args: &args{
				table:          mockTable,
				matchFieldName: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := ts.getMatchFieldByName(tt.args.table, tt.args.matchFieldName)
				if tt.wantErr {
					require.Nil(t, got)
					return
				}
				require.Equal(t, tt.want.name, got.GetName())
				require.Equal(t, tt.want.id, got.GetId())
				require.Equal(t, tt.want.match, got.GetMatchType())
			},
		)
	}
}

func Test_getActionbyID(t *testing.T) {
	type args struct {
		actionID   uint32
		translator *P4rtTranslator
	}
	type want struct {
		actionId   uint32
		actionName string
	}

	tests := []struct {
		name    string
		args    *args
		want    *want
		wantErr bool
	}{
		{name: "get existing action",
			args: &args{
				actionID:   32742981,
				translator: setupNewTranslator(t),
			},
			want: &want{actionName: "PreQosPipe.load_tunnel_param",
				actionId: 32742981,
			},
		},
		{name: "non existing action",
			args: &args{
				actionID:   9999999,
				translator: setupNewTranslator(t),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.args.translator.getActionByID(tt.args.actionID)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.Equal(t, tt.want.actionName, got.GetPreamble().GetName())
				require.Equal(t, tt.want.actionId, got.GetPreamble().GetId())
			},
		)
	}
}

func Test_getActionParamByName(t *testing.T) {
	ts := setupNewTranslator(t)

	mockAction, err := ts.getActionByID(32742981) //PreQosPipe.load_tunnel_param
	require.NoError(t, err)

	type args struct {
		action          *p4ConfigV1.Action
		actionParamName string
	}

	type want struct {
		paramName string
		paramId   uint32
		bitwidth  int32
	}

	tests := []struct {
		name    string
		args    *args
		want    *want
		wantErr bool
	}{
		{name: "Existing action param name",
			args: &args{
				action:          mockAction,
				actionParamName: "src_addr",
			},
			want: &want{paramName: "src_addr",
				paramId:  1,
				bitwidth: 32,
			},
		},
		{name: "non existing action param name",
			args: &args{
				action:          mockAction,
				actionParamName: "test",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := ts.getActionParamByName(tt.args.action, tt.args.actionParamName)
				if tt.wantErr {
					require.Nil(t, got)
					return
				}
				require.Equal(t, tt.want.paramName, got.GetName())
				require.Equal(t, tt.want.paramId, got.GetId())
				require.Equal(t, tt.want.bitwidth, got.GetBitwidth())
			},
		)
	}
}

func Test_withLPMField(t *testing.T) {
	ipString := "172.16.0.1"

	octets := strings.Split(ipString, ".")
	octet0, _ := strconv.Atoi(octets[0])
	octet1, _ := strconv.Atoi(octets[1])
	octet2, _ := strconv.Atoi(octets[2])
	octet3, _ := strconv.Atoi(octets[3])

	ipBytes := []byte{byte(octet0), byte(octet1), byte(octet2), byte(octet3)}

	ipToUint32 := func(ip string) uint32 {
		var long uint32
		err := binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long)
		require.NoError(t, err)

		return long
	}

	type args struct {
		tableEntry   *p4.TableEntry
		lpmName      string
		value        uint32
		prefixLength uint8
	}

	type want struct {
		match      *p4.FieldMatch_LPM
		numMatches int
	}

	tests := []struct {
		name       string
		args       *args
		translator *P4rtTranslator
		want       *want
		wantErr    bool
	}{
		{name: "Add LPMMatch",
			args: &args{
				tableEntry: &p4.TableEntry{
					TableId:  39015874,
					Priority: DefaultPriority,
				},
				lpmName:      "dst_prefix",
				value:        ipToUint32(ipString),
				prefixLength: uint8(16),
			},
			translator: setupNewTranslator(t),
			want: &want{
				match: &p4.FieldMatch_LPM{
					Value:                ipBytes,
					PrefixLen:            16,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     []uint8(nil),
					XXX_sizecache:        0,
				},
				numMatches: 1,
			},
		},
		{name: "non existent parameter name",
			args: &args{
				tableEntry: &p4.TableEntry{
					TableId:  39015874,
					Priority: DefaultPriority,
				},
				lpmName:      "test",
				value:        ipToUint32(ipString),
				prefixLength: uint8(16),
			},
			translator: setupNewTranslator(t),
			want:       &want{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := tt.translator.withLPMField(tt.args.tableEntry, tt.args.lpmName, tt.args.value, tt.args.prefixLength)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.Equal(t, tt.want.numMatches, len(tt.args.tableEntry.GetMatch()))
				require.Equal(t, tt.want.match, tt.args.tableEntry.GetMatch()[0].GetLpm())
			},
		)
	}
}
