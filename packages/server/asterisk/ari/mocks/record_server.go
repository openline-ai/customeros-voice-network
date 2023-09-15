package mocks

import "github.com/stretchr/testify/mock"

type RtpServerMock struct {
	mock.Mock
}

func (fsc RtpServerMock) Address() string {
	args := fsc.Called()
	return args.String(0)
}

func (fsc RtpServerMock) Close() {
	fsc.Called()
}

func (fsc RtpServerMock) Listen() error {
	args := fsc.Called()
	return args.Error(0)
}
