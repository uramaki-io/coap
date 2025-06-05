# COAP Client/Server Library

[![Go Reference](https://pkg.go.dev/badge/github.com/uramaki-io/coap.svg)](https://pkg.go.dev/github.com/uramaki-io/coap)
[![codecov](https://codecov.io/gh/uramaki-io/coap/graph/badge.svg?token=NEP6SSQ8MB)](https://codecov.io/gh/uramaki-io/coap)

## Purpose

Implementation of the Constrained Application Protocol (CoAP), designed to facilitate lightweight and efficient communication for resource-constrained devices in Internet of Things (IoT) and machine-to-machine (M2M) applications.

Its primary purpose is to provide developers with a performant and easy-to-use toolkit for building CoAP-based clients and servers in Go, adhering to the [RFC 7252](https://datatracker.ietf.org/doc/html/rfc7252) specification.

The library supports **UDP** and **DTLS** protocols.

## Structure

The `Message` type represents a complete CoAP message, encapsulating the protocol's header, options, and payload.

