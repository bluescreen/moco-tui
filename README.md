# Moco Time Tracking CLI

A command-line interface for tracking time entries in Moco, built with Go and Bubble Tea.

## Prerequisites

- Go 1.21.1 or later
- A Moco account with API access

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd moco-golang
```

2. Copy the example environment file and update it with your credentials:
```bash
cp .env.example .env
```
Then edit the `.env` file with your Moco API credentials:
```bash
MOCO_API_KEY=your_api_key_here
MOCO_DOMAIN=your_domain_here
```

3. Install dependencies:
```bash
go mod download
```

## Running the Application

### Development Mode

From the `src` directory:
```bash
go run ./src
```

## Features

- View time entries in a table format
- Add new time entries
- Filter and search time entries
- Interactive command-line interface





