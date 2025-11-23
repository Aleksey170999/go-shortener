// Package repository provides interfaces and implementations for URL storage and retrieval.
// It includes both in-memory and database-backed implementations of the URLRepository interface.
//
// The main interface is URLRepository which defines the contract for URL storage operations.
// Two implementations are provided:
// - memoryURLRepository: In-memory storage using a map
// - DataBaseURLRepository: Persistent storage using PostgreSQL
package repository
