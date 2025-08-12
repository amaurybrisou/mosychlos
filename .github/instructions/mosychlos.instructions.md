---
applyTo: '**/*'
---

**IMPORTANT**

At all costs, every logical entity in the app should receive shared bag either through constructor injection or method parameters.

Never write to types having New constructor inside the same file

When you recreate a file:

- write the new file with \_new suffix
- override the old one with the new one by using mv command

```bash
# assuming we have a file named: myfile.go
# create the new file:  myfile_new.go
# then :
mv myfile_new.go myfile.go

# Mosychlos Development Instructions Index

This is the master instruction file for the Mosychlos portfolio management system. It provides comprehensive guidelines for Go CLI development with institutional-grade financial analysis capabilities.

## **Project Overview**

Mosychlos is a sophisticated portfolio management and analysis system comprising:

- **Go CLI Tool**: Core portfolio management, data processing, and command-line interface
- **Engine Architecture**: Tool-driven multi-engine analysis pipeline with professional roles
- **Financial Integrations**: FMP, FRED, Binance, NewsAPI, and OpenAI services
- **Professional Reports**: PDF generation, charts, and comprehensive financial analysis

## **Project Tree**

[Tree](tree.md)

## **Architecture Components**

```

┌─────────────────────────────────────────────────────────────────┐
│ Mosychlos CLI Architecture │
│ │
│ ┌─────────────────┐ ┌──────────────────────────────────┐ │
│ │ CLI Interface │────│ Engine Orchestrator │ │
│ │ │ │ │ │
│ │ • Commands │ │ • Sequential Engine Execution │ │
│ │ • Reports │ │ • SharedBag Context Management │ │
│ │ • Validation │ │ • Tool Constraint Enforcement │ │
│ └─────────────────┘ └──────────────────────────────────┘ │
│ │ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Professional Engines │ │
│ │ │ │
│ │ ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐ │ │
│ │ │ Financial │ │ Risk │ │ Investment │ │ │
│ │ │ Analysis │ │ Analysis │ │ Committee │ │ │
│ │ │ Engine │ │ Engine │ │ Engine │ │ │
│ │ └─────────────┘ └─────────────┘ └─────────────────────┘ │ │
│ └─────────────────────────────────────────────────────────┘ │
│ │ │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Tool Ecosystem │ │
│ │ │ │
│ │ • FMP (Financial Data) • FRED (Economic Data) │ │
│ │ • NewsAPI (Market News) • Weather (Conditions) │ │
│ │ • OpenAI (Analysis) • Binance (Crypto) │ │
│ │ • Cached Tools • Metrics Wrapper │ │
│ └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘

````

## **Development Guidelines**

### 🔧 **[Core Rules](core.rules.md)**

Fundamental development principles, commit standards, and code quality rules.

### 🧪 **[Testing](testing.md)**

Table-driven tests, mock strategies (using mockgen only), and test suite patterns for reliable code coverage.

### 📝 **[Logging](logging.md)**

Structured logging standards with essential context, performance tracking, and security considerations.

### ✅ **[Validation](validation.md)**

Input validation rules, business logic checks, and error handling patterns.

### 🔨 **[Build](build.md)**

Build processes, compilation guidelines, and deployment procedures.

### 🔍 **[Linting](linting.md)**

Code quality standards, static analysis rules, and formatting requirements.

### 🚀 **[Running](run.md)**

Execution commands, usage examples, and operational procedures.

```bash
# Run specific tools
go run . analyze --portfolio=current
go run . tools list
````

### 📋 **[Go Language Guidelines](golang.instructions.md)**

Comprehensive Go development patterns, external service integration, and architecture-specific coding standards.

### 🔧 **[Tool Generation](tool-generation.md)**

Step-by-step guide for creating new tools in the tools ecosystem with caching, monitoring, and proper integration.

## **Critical Development Principles**

### **Tool-Driven Architecture**

- Every analysis capability is implemented as a tool with proper constraints
- Professional behavior emerges from tool access patterns and usage limits
- Engine chaining provides single-context multi-agent capabilities

### **Configuration Management**

- Never use hardcoded values or mocks (except mockgen)
- Configuration should break runtime when invalid
- Centralized localization drives geographic and regional settings

### **Engine Implementation**

- Always respect the 4-layer rules in modules
- Use existing common modules to avoid duplicating implementations
- Never keep legacy code - always patch existing implementations

### **Caching Strategy**

- Always think about caching for expensive operations
- Use cached tools for external API calls
- Implement proper cache invalidation and monitoring

### **Integration Patterns**

- Use HTTP clients for external service integration
- Implement comprehensive error handling with proper logging
- Ensure we can send every parameter external services expect

## **Quality Standards**

### **Testing Requirements**

- Maintain >80% test coverage
- Always run tests with `-race` flag
- Use table-driven test patterns consistently
- Mock all external interfaces

### **Code Quality**

- Pass all linters (`golangci-lint`)
- Follow Conventional Commits standard
- Use descriptive branch names: `feat/`, `fix/`, `chore/`
- Add comments only when they provide real value

### **Performance Requirements**

- API responses should complete within 30 seconds
- Support multiple concurrent analysis requests
- Cache expensive computations appropriately
- Monitor and manage resource usage

## **External Resources**

### **Prompt Engineering Best Practices**

Follow GitHub Copilot's prompt engineering guidelines for effective AI assistance:

- 📖 [GitHub Copilot Prompt Engineering](https://docs.github.com/en/copilot/concepts/prompt-engineering)

### **Go Development Resources**

- 📘 [Cobra CLI Guidelines](https://github.com/spf13/cobra/blob/master/docs/_index.md)
- 📘 [Go Modules Documentation](https://blog.golang.org/using-go-modules)
- 📘 [Effective Go](https://golang.org/doc/effective_go.html)
- 📘 [Conventional Commits](https://www.conventionalcommits.org/)

## **Getting Started**

1. **📚 Read Core Instructions**: Start with [core.rules.md](core.rules.md) for fundamental principles
2. **🛠️ Setup Environment**: Follow [build.md](build.md) for development setup
3. **🧪 Run Tests**: Use [testing.md](testing.md) patterns for reliable code
4. **📝 Follow Standards**: Apply [logging.md](logging.md) and [validation.md](validation.md) guidelines
5. **🚀 Execute**: Use [run.md](run.md) for operational commands

## **Support and Resources**

- **Code Quality**: Follow [linting.md](linting.md) standards during development
- **Issue Resolution**: Consult specific instruction files for troubleshooting guidance
- **Best Practices**: Leverage external resources and established patterns
- **Architecture Help**: Review engine diagrams and tool integration documentation

---
