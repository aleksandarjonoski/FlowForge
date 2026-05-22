# FlowForge

FlowForge is a visual integration and workflow engine for Go.

It allows developers to design systems visually by connecting nodes such as APIs, databases, queues, and transformations, then executes them using a lightweight Go runtime.

## 🚀 Vision

Modern integration platforms are either:
- too enterprise-heavy (TIBCO-style systems)
- too limited (simple automation tools)
- or not developer-friendly

FlowForge aims to bridge this gap by combining:
- visual workflow design
- Go-based execution engine
- extensible node/plugin system
- lightweight, self-hostable runtime

## 🧩 Core Concept

Everything in FlowForge is a graph:

- Nodes = actions (HTTP, DB, Kafka, transform, etc.)
- Edges = data flow between nodes
- Workflow = executable directed graph

Workflows are defined as JSON and executed by a Go runtime.

## ⚙️ Current Status

This project is in early development.

Currently working on:
- Visual workflow editor (frontend)
- Node system design
- JSON-based flow model

Not yet implemented:
- Runtime execution engine
- Connectors (Kafka, DB, APIs)
- Code generation
- Distributed execution

## 🏗️ Architecture (High Level)

UI (Visual Builder)

        ↓
        
Flow JSON Definition

        ↓
        
Go Execution Engine

        ↓
        
Integration Runtime

## 📌 Goals

- Make integration systems easy to build visually
- Keep runtime fast, lightweight, and Go-native
- Allow plugins for custom nodes and connectors
- Enable self-hosted deployment

## 💡 Inspiration

Inspired by:
- Node-RED
- Apache NiFi
- TIBCO BusinessWorks
- MuleSoft
- n8n

But redesigned for Go-first execution and modern developer workflows.

## ⚠️ Disclaimer

This project is experimental and under active development.
Expect breaking changes.

---
