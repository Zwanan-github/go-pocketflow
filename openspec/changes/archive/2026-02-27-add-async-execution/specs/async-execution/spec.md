## ADDED Requirements

### Requirement: AsyncNode supports asynchronous execution
The system SHALL provide AsyncNode that executes node phases asynchronously using goroutines.

#### Scenario: AsyncNode executes Prep phase asynchronously
- **WHEN** AsyncNode.Prep is called with an async function
- **THEN** system SHALL execute the Prep function in a separate goroutine
- **AND** SHALL return a channel for receiving the result

#### Scenario: AsyncNode executes Exec phase asynchronously
- **WHEN** AsyncNode.Exec is called with an async function
- **THEN** system SHALL execute the Exec function in a separate goroutine
- **AND** SHALL return a channel for receiving the result

#### Scenario: AsyncNode executes Post phase asynchronously
- **WHEN** AsyncNode.Post is called with an async function
- **THEN** system SHALL execute the Post function in a separate goroutine
- **AND** SHALL return a channel for receiving the action result

### Requirement: AsyncNode provides concurrency control
The system SHALL allow limiting the number of concurrent AsyncNode executions.

#### Scenario: Concurrency limit enforced
- **WHEN** AsyncNode is configured with max concurrency of 5
- **AND** 10 AsyncNodes are started simultaneously
- **THEN** system SHALL execute maximum 5 nodes concurrently
- **AND** SHALL queue remaining nodes until slots become available

#### Scenario: Dynamic concurrency adjustment
- **WHEN** AsyncNode concurrency limit is changed during execution
- **THEN** system SHALL apply new limit to new executions
- **AND** SHALL not interrupt already running executions

### Requirement: AsyncNode supports timeout control
The system SHALL provide timeout mechanism for AsyncNode execution.

#### Scenario: Execution timeout triggers cancellation
- **WHEN** AsyncNode execution exceeds configured timeout
- **THEN** system SHALL cancel the execution
- **AND** SHALL return timeout error
- **AND** SHALL clean up goroutine resources

#### Scenario: Timeout per phase
- **WHEN** AsyncNode is configured with different timeouts for Prep, Exec, and Post phases
- **THEN** system SHALL apply timeout independently to each phase
- **AND** SHALL cancel only the phase that exceeds timeout

### Requirement: AsyncNode handles errors asynchronously
The system SHALL provide error propagation mechanism for asynchronous executions.

#### Scenario: Error in async execution
- **WHEN** AsyncNode execution encounters an error
- **THEN** system SHALL propagate error through result channel
- **AND** SHALL not block other concurrent executions

#### Scenario: Error recovery with retry
- **WHEN** AsyncNode execution fails and retry is configured
- **THEN** system SHALL automatically retry the execution
- **AND** SHALL respect retry limit configuration