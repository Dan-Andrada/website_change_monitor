# Website Change Monitor

## 1. Project Description

The **Website Change Monitor** is a command-line application developed in Go, designed to periodically analyze the content of specified webpages and detect updates. The system enables users to register webpages for monitoring, define the HTML elements of interest through CSS selectors, configure checking intervals, and receive notifications when relevant changes are identified.

The project integrates essential concepts such as web scraping, scheduling of automated tasks, hashing and change detection, storage of historical data, and basic CLI-based user interaction. The purpose of the application is to offer a lightweight and extensible tool capable of monitoring dynamic online content in an efficient and automated manner.

---

## 2. Use Case Diagram

The main functionalities of the system are represented in the following use case diagram:

![Use Case Diagram](docs/usecase.png)

---

## 3. Running the Application

### 3.1 Prerequisites

- Go installed on the system
- Google Chrome installed (required for screenshot capture)
- Internet access

---

### 3.2 Build

Before running the application, it must be compiled:

```bash
go build -o gomonitor.exe
```

---

## 3. Running the Application

### 3.1 Prerequisites

- Go installed on the system
- Google Chrome installed (required for screenshot capture)
- Internet access

---

### 3.2 Build

Before running the application, it must be compiled:

```bash
go build -o gomonitor.exe
```

---

### 3.3 CLI Commands

The application is fully controlled through command-line arguments and supports the following commands:

#### Add a new monitored page

```bash
gomonitor.exe add <url> <css_selector> <frequency_minutes>
```

#### Example:

```bash
gomonitor.exe add "http://127.0.0.1:5500/test.html" "#price" 1
```

This command registers a new webpage for monitoring by specifying:

- the URL
- the CSS selector of the monitored element
- the checking interval in minutes

The configuration is stored in data/monitors.json.

#### Run the monitoring process

```bash
gomonitor.exe run
```

This command starts the monitoring service. All registered webpages are checked periodically according to their configured frequency. When a change is detected, the application:

- computes a content hash

- generates a textual diff

- captures before and after screenshots

- saves the change in history

- sends a notification

The process runs continuously until manually stopped.

---

## 4. Output Files and Directories

During execution, the application generates and updates the following files and folders:

data/monitors.json – stores monitored URLs and change history

screenshots/ – contains before and after screenshots for each detected change

monitor.log – execution logs and detected events




