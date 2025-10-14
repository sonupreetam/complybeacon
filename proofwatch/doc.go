// Package proofwatch provides a comprehensive policy evidence logging and monitoring system
// built on OpenTelemetry. It enables organizations to track, log, and monitor policy
// evaluation events and compliance evidence in real-time.
//
// Basic Usage:
//
//	// Create a new ProofWatch instance
//	pw, err := proofwatch.NewProofWatch()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Log evidence with default severity
//	err = pw.Log(ctx, evidence)
//
//	// Log evidence with specific severity
//	err = pw.LogWithSeverity(ctx, evidence, olog.SeverityWarn)
//
// Configuration:
//
//	// Create with custom providers
//	pw, err := proofwatch.NewProofWatch(
//		proofwatch.WithLoggerProvider(customLoggerProvider),
//		proofwatch.WithMeterProvider(customMeterProvider),
//		proofwatch.WithTracerProvider(customTracerProvider),
//	)
//
// Metrics:
//   - evidence_processed_count: Total number of evidence items processed successfully
//   - evidence_dropped_count: Total number of evidence items dropped due to failures
//
// Traces:
//   - evidence.log_evidence: Tracks the complete evidence logging process
//   - evidence.logged: Event marker for successful evidence logging
package proofwatch
