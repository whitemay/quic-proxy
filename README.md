# QUIC Proxy

A bridge from QUIC to TCP.

## Overview

This project provides a proxy server that listens for QUIC connections and forwards the traffic to a TCP server. It is designed to facilitate communication between QUIC-based clients and traditional TCP-based services.

For Chinese documentation, see [README_zh.md](README_zh.md).

## Features

- Listens for QUIC connections and forwards data to a TCP server.
- Supports environment variable configuration for port, certificate files, and more.
- Includes a Dockerfile for easy cross-platform builds and deployment.

## Prerequisites

- Go 1.20 or higher
- QUIC-compatible certificates (e.g., generated using OpenSSL)

## Installation

1. Clone the repository:
2. Create a .env file in the current directory based on .env.example, and edit it accordingly.

   Run the project:

   ```bash
   go get
   go run .
   ```
3. Build the Docker image:

   ```bash
   docker build -t quic-proxy .
   docker run -d -p 443:443 quic-proxy
   ```