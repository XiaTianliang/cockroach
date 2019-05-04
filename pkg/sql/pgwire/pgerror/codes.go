// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package pgerror

import "github.com/cockroachdb/cockroach/pkg/util/pgcode"

// PG error codes as defined by the pgcode package.
//
// These forward definitions are introduced so as to not require to
// update the entire SQL codebase at the same time as the introduction
// of the errors package.
// They can be removed at a later stage. New code should use the
// pgcode package directly.
const (
	CodeSuccessfulCompletionError                              = pgcode.SuccessfulCompletion
	CodeWarningError                                           = pgcode.Warning
	CodeWarningDynamicResultSetsReturnedError                  = pgcode.WarningDynamicResultSetsReturned
	CodeWarningImplicitZeroBitPaddingError                     = pgcode.WarningImplicitZeroBitPadding
	CodeWarningNullValueEliminatedInSetFunctionError           = pgcode.WarningNullValueEliminatedInSetFunction
	CodeWarningPrivilegeNotGrantedError                        = pgcode.WarningPrivilegeNotGranted
	CodeWarningPrivilegeNotRevokedError                        = pgcode.WarningPrivilegeNotRevoked
	CodeWarningStringDataRightTruncationError                  = pgcode.WarningStringDataRightTruncation
	CodeWarningDeprecatedFeatureError                          = pgcode.WarningDeprecatedFeature
	CodeNoDataError                                            = pgcode.NoData
	CodeNoAdditionalDynamicResultSetsReturnedError             = pgcode.NoAdditionalDynamicResultSetsReturned
	CodeSQLStatementNotYetCompleteError                        = pgcode.SQLStatementNotYetComplete
	CodeConnectionExceptionError                               = pgcode.ConnectionException
	CodeConnectionDoesNotExistError                            = pgcode.ConnectionDoesNotExist
	CodeConnectionFailureError                                 = pgcode.ConnectionFailure
	CodeSQLclientUnableToEstablishSQLconnectionError           = pgcode.SQLclientUnableToEstablishSQLconnection
	CodeSQLserverRejectedEstablishmentOfSQLconnectionError     = pgcode.SQLserverRejectedEstablishmentOfSQLconnection
	CodeTransactionResolutionUnknownError                      = pgcode.TransactionResolutionUnknown
	CodeProtocolViolationError                                 = pgcode.ProtocolViolation
	CodeTriggeredActionExceptionError                          = pgcode.TriggeredActionException
	CodeFeatureNotSupportedError                               = pgcode.FeatureNotSupported
	CodeInvalidTransactionInitiationError                      = pgcode.InvalidTransactionInitiation
	CodeLocatorExceptionError                                  = pgcode.LocatorException
	CodeInvalidLocatorSpecificationError                       = pgcode.InvalidLocatorSpecification
	CodeInvalidGrantorError                                    = pgcode.InvalidGrantor
	CodeInvalidGrantOperationError                             = pgcode.InvalidGrantOperation
	CodeInvalidRoleSpecificationError                          = pgcode.InvalidRoleSpecification
	CodeDiagnosticsExceptionError                              = pgcode.DiagnosticsException
	CodeStackedDiagnosticsAccessedWithoutActiveHandlerError    = pgcode.StackedDiagnosticsAccessedWithoutActiveHandler
	CodeCaseNotFoundError                                      = pgcode.CaseNotFound
	CodeCardinalityViolationError                              = pgcode.CardinalityViolation
	CodeDataExceptionError                                     = pgcode.DataException
	CodeArraySubscriptError                                    = pgcode.ArraySubscript
	CodeCharacterNotInRepertoireError                          = pgcode.CharacterNotInRepertoire
	CodeDatetimeFieldOverflowError                             = pgcode.DatetimeFieldOverflow
	CodeDivisionByZeroError                                    = pgcode.DivisionByZero
	CodeInvalidWindowFrameOffsetError                          = pgcode.InvalidWindowFrameOffset
	CodeErrorInAssignmentError                                 = pgcode.ErrorInAssignment
	CodeEscapeCharacterConflictError                           = pgcode.EscapeCharacterConflict
	CodeIndicatorOverflowError                                 = pgcode.IndicatorOverflow
	CodeIntervalFieldOverflowError                             = pgcode.IntervalFieldOverflow
	CodeInvalidArgumentForLogarithmError                       = pgcode.InvalidArgumentForLogarithm
	CodeInvalidArgumentForNtileFunctionError                   = pgcode.InvalidArgumentForNtileFunction
	CodeInvalidArgumentForNthValueFunctionError                = pgcode.InvalidArgumentForNthValueFunction
	CodeInvalidArgumentForPowerFunctionError                   = pgcode.InvalidArgumentForPowerFunction
	CodeInvalidArgumentForWidthBucketFunctionError             = pgcode.InvalidArgumentForWidthBucketFunction
	CodeInvalidCharacterValueForCastError                      = pgcode.InvalidCharacterValueForCast
	CodeInvalidDatetimeFormatError                             = pgcode.InvalidDatetimeFormat
	CodeInvalidEscapeCharacterError                            = pgcode.InvalidEscapeCharacter
	CodeInvalidEscapeOctetError                                = pgcode.InvalidEscapeOctet
	CodeInvalidEscapeSequenceError                             = pgcode.InvalidEscapeSequence
	CodeNonstandardUseOfEscapeCharacterError                   = pgcode.NonstandardUseOfEscapeCharacter
	CodeInvalidIndicatorParameterValueError                    = pgcode.InvalidIndicatorParameterValue
	CodeInvalidParameterValueError                             = pgcode.InvalidParameterValue
	CodeInvalidRegularExpressionError                          = pgcode.InvalidRegularExpression
	CodeInvalidRowCountInLimitClauseError                      = pgcode.InvalidRowCountInLimitClause
	CodeInvalidRowCountInResultOffsetClauseError               = pgcode.InvalidRowCountInResultOffsetClause
	CodeInvalidTimeZoneDisplacementValueError                  = pgcode.InvalidTimeZoneDisplacementValue
	CodeInvalidUseOfEscapeCharacterError                       = pgcode.InvalidUseOfEscapeCharacter
	CodeMostSpecificTypeMismatchError                          = pgcode.MostSpecificTypeMismatch
	CodeNullValueNotAllowedError                               = pgcode.NullValueNotAllowed
	CodeNullValueNoIndicatorParameterError                     = pgcode.NullValueNoIndicatorParameter
	CodeNumericValueOutOfRangeError                            = pgcode.NumericValueOutOfRange
	CodeSequenceGeneratorLimitExceeded                         = pgcode.SequenceGeneratorLimitExceeded
	CodeStringDataLengthMismatchError                          = pgcode.StringDataLengthMismatch
	CodeStringDataRightTruncationError                         = pgcode.StringDataRightTruncation
	CodeSubstringError                                         = pgcode.Substring
	CodeTrimError                                              = pgcode.Trim
	CodeUnterminatedCStringError                               = pgcode.UnterminatedCString
	CodeZeroLengthCharacterStringError                         = pgcode.ZeroLengthCharacterString
	CodeFloatingPointExceptionError                            = pgcode.FloatingPointException
	CodeInvalidTextRepresentationError                         = pgcode.InvalidTextRepresentation
	CodeInvalidBinaryRepresentationError                       = pgcode.InvalidBinaryRepresentation
	CodeBadCopyFileFormatError                                 = pgcode.BadCopyFileFormat
	CodeUntranslatableCharacterError                           = pgcode.UntranslatableCharacter
	CodeNotAnXMLDocumentError                                  = pgcode.NotAnXMLDocument
	CodeInvalidXMLDocumentError                                = pgcode.InvalidXMLDocument
	CodeInvalidXMLContentError                                 = pgcode.InvalidXMLContent
	CodeInvalidXMLCommentError                                 = pgcode.InvalidXMLComment
	CodeInvalidXMLProcessingInstructionError                   = pgcode.InvalidXMLProcessingInstruction
	CodeIntegrityConstraintViolationError                      = pgcode.IntegrityConstraintViolation
	CodeRestrictViolationError                                 = pgcode.RestrictViolation
	CodeNotNullViolationError                                  = pgcode.NotNullViolation
	CodeForeignKeyViolationError                               = pgcode.ForeignKeyViolation
	CodeUniqueViolationError                                   = pgcode.UniqueViolation
	CodeCheckViolationError                                    = pgcode.CheckViolation
	CodeExclusionViolationError                                = pgcode.ExclusionViolation
	CodeInvalidCursorStateError                                = pgcode.InvalidCursorState
	CodeInvalidTransactionStateError                           = pgcode.InvalidTransactionState
	CodeActiveSQLTransactionError                              = pgcode.ActiveSQLTransaction
	CodeBranchTransactionAlreadyActiveError                    = pgcode.BranchTransactionAlreadyActive
	CodeHeldCursorRequiresSameIsolationLevelError              = pgcode.HeldCursorRequiresSameIsolationLevel
	CodeInappropriateAccessModeForBranchTransactionError       = pgcode.InappropriateAccessModeForBranchTransaction
	CodeInappropriateIsolationLevelForBranchTransactionError   = pgcode.InappropriateIsolationLevelForBranchTransaction
	CodeNoActiveSQLTransactionForBranchTransactionError        = pgcode.NoActiveSQLTransactionForBranchTransaction
	CodeReadOnlySQLTransactionError                            = pgcode.ReadOnlySQLTransaction
	CodeSchemaAndDataStatementMixingNotSupportedError          = pgcode.SchemaAndDataStatementMixingNotSupported
	CodeNoActiveSQLTransactionError                            = pgcode.NoActiveSQLTransaction
	CodeInFailedSQLTransactionError                            = pgcode.InFailedSQLTransaction
	CodeInvalidSQLStatementNameError                           = pgcode.InvalidSQLStatementName
	CodeTriggeredDataChangeViolationError                      = pgcode.TriggeredDataChangeViolation
	CodeInvalidAuthorizationSpecificationError                 = pgcode.InvalidAuthorizationSpecification
	CodeInvalidPasswordError                                   = pgcode.InvalidPassword
	CodeDependentPrivilegeDescriptorsStillExistError           = pgcode.DependentPrivilegeDescriptorsStillExist
	CodeDependentObjectsStillExistError                        = pgcode.DependentObjectsStillExist
	CodeInvalidTransactionTerminationError                     = pgcode.InvalidTransactionTermination
	CodeSQLRoutineExceptionError                               = pgcode.SQLRoutineException
	CodeRoutineExceptionFunctionExecutedNoReturnStatementError = pgcode.RoutineExceptionFunctionExecutedNoReturnStatement
	CodeRoutineExceptionModifyingSQLDataNotPermittedError      = pgcode.RoutineExceptionModifyingSQLDataNotPermitted
	CodeRoutineExceptionProhibitedSQLStatementAttemptedError   = pgcode.RoutineExceptionProhibitedSQLStatementAttempted
	CodeRoutineExceptionReadingSQLDataNotPermittedError        = pgcode.RoutineExceptionReadingSQLDataNotPermitted
	CodeInvalidCursorNameError                                 = pgcode.InvalidCursorName
	CodeExternalRoutineExceptionError                          = pgcode.ExternalRoutineException
	CodeExternalRoutineContainingSQLNotPermittedError          = pgcode.ExternalRoutineContainingSQLNotPermitted
	CodeExternalRoutineModifyingSQLDataNotPermittedError       = pgcode.ExternalRoutineModifyingSQLDataNotPermitted
	CodeExternalRoutineProhibitedSQLStatementAttemptedError    = pgcode.ExternalRoutineProhibitedSQLStatementAttempted
	CodeExternalRoutineReadingSQLDataNotPermittedError         = pgcode.ExternalRoutineReadingSQLDataNotPermitted
	CodeExternalRoutineInvocationExceptionError                = pgcode.ExternalRoutineInvocationException
	CodeExternalRoutineInvalidSQLstateReturnedError            = pgcode.ExternalRoutineInvalidSQLstateReturned
	CodeExternalRoutineNullValueNotAllowedError                = pgcode.ExternalRoutineNullValueNotAllowed
	CodeExternalRoutineTriggerProtocolViolatedError            = pgcode.ExternalRoutineTriggerProtocolViolated
	CodeExternalRoutineSrfProtocolViolatedError                = pgcode.ExternalRoutineSrfProtocolViolated
	CodeSavepointExceptionError                                = pgcode.SavepointException
	CodeInvalidSavepointSpecificationError                     = pgcode.InvalidSavepointSpecification
	CodeInvalidCatalogNameError                                = pgcode.InvalidCatalogName
	CodeInvalidSchemaNameError                                 = pgcode.InvalidSchemaName
	CodeTransactionRollbackError                               = pgcode.TransactionRollback
	CodeTransactionIntegrityConstraintViolationError           = pgcode.TransactionIntegrityConstraintViolation
	CodeSerializationFailureError                              = pgcode.SerializationFailure
	CodeStatementCompletionUnknownError                        = pgcode.StatementCompletionUnknown
	CodeDeadlockDetectedError                                  = pgcode.DeadlockDetected
	CodeSyntaxErrorOrAccessRuleViolationError                  = pgcode.SyntaxErrorOrAccessRuleViolation
	CodeSyntaxError                                            = pgcode.Syntax
	CodeInsufficientPrivilegeError                             = pgcode.InsufficientPrivilege
	CodeCannotCoerceError                                      = pgcode.CannotCoerce
	CodeGroupingError                                          = pgcode.Grouping
	CodeWindowingError                                         = pgcode.Windowing
	CodeInvalidRecursionError                                  = pgcode.InvalidRecursion
	CodeInvalidForeignKeyError                                 = pgcode.InvalidForeignKey
	CodeInvalidNameError                                       = pgcode.InvalidName
	CodeNameTooLongError                                       = pgcode.NameTooLong
	CodeReservedNameError                                      = pgcode.ReservedName
	CodeDatatypeMismatchError                                  = pgcode.DatatypeMismatch
	CodeIndeterminateDatatypeError                             = pgcode.IndeterminateDatatype
	CodeCollationMismatchError                                 = pgcode.CollationMismatch
	CodeIndeterminateCollationError                            = pgcode.IndeterminateCollation
	CodeWrongObjectTypeError                                   = pgcode.WrongObjectType
	CodeUndefinedColumnError                                   = pgcode.UndefinedColumn
	CodeUndefinedFunctionError                                 = pgcode.UndefinedFunction
	CodeUndefinedTableError                                    = pgcode.UndefinedTable
	CodeUndefinedParameterError                                = pgcode.UndefinedParameter
	CodeUndefinedObjectError                                   = pgcode.UndefinedObject
	CodeDuplicateColumnError                                   = pgcode.DuplicateColumn
	CodeDuplicateCursorError                                   = pgcode.DuplicateCursor
	CodeDuplicateDatabaseError                                 = pgcode.DuplicateDatabase
	CodeDuplicateFunctionError                                 = pgcode.DuplicateFunction
	CodeDuplicatePreparedStatementError                        = pgcode.DuplicatePreparedStatement
	CodeDuplicateSchemaError                                   = pgcode.DuplicateSchema
	CodeDuplicateRelationError                                 = pgcode.DuplicateRelation
	CodeDuplicateAliasError                                    = pgcode.DuplicateAlias
	CodeDuplicateObjectError                                   = pgcode.DuplicateObject
	CodeAmbiguousColumnError                                   = pgcode.AmbiguousColumn
	CodeAmbiguousFunctionError                                 = pgcode.AmbiguousFunction
	CodeAmbiguousParameterError                                = pgcode.AmbiguousParameter
	CodeAmbiguousAliasError                                    = pgcode.AmbiguousAlias
	CodeInvalidColumnReferenceError                            = pgcode.InvalidColumnReference
	CodeInvalidColumnDefinitionError                           = pgcode.InvalidColumnDefinition
	CodeInvalidCursorDefinitionError                           = pgcode.InvalidCursorDefinition
	CodeInvalidDatabaseDefinitionError                         = pgcode.InvalidDatabaseDefinition
	CodeInvalidFunctionDefinitionError                         = pgcode.InvalidFunctionDefinition
	CodeInvalidPreparedStatementDefinitionError                = pgcode.InvalidPreparedStatementDefinition
	CodeInvalidSchemaDefinitionError                           = pgcode.InvalidSchemaDefinition
	CodeInvalidTableDefinitionError                            = pgcode.InvalidTableDefinition
	CodeInvalidObjectDefinitionError                           = pgcode.InvalidObjectDefinition
	CodeWithCheckOptionViolationError                          = pgcode.WithCheckOptionViolation
	CodeInsufficientResourcesError                             = pgcode.InsufficientResources
	CodeDiskFullError                                          = pgcode.DiskFull
	CodeOutOfMemoryError                                       = pgcode.OutOfMemory
	CodeTooManyConnectionsError                                = pgcode.TooManyConnections
	CodeConfigurationLimitExceededError                        = pgcode.ConfigurationLimitExceeded
	CodeProgramLimitExceededError                              = pgcode.ProgramLimitExceeded
	CodeStatementTooComplexError                               = pgcode.StatementTooComplex
	CodeTooManyColumnsError                                    = pgcode.TooManyColumns
	CodeTooManyArgumentsError                                  = pgcode.TooManyArguments
	CodeObjectNotInPrerequisiteStateError                      = pgcode.ObjectNotInPrerequisiteState
	CodeObjectInUseError                                       = pgcode.ObjectInUse
	CodeCantChangeRuntimeParamError                            = pgcode.CantChangeRuntimeParam
	CodeLockNotAvailableError                                  = pgcode.LockNotAvailable
	CodeOperatorInterventionError                              = pgcode.OperatorIntervention
	CodeQueryCanceledError                                     = pgcode.QueryCanceled
	CodeAdminShutdownError                                     = pgcode.AdminShutdown
	CodeCrashShutdownError                                     = pgcode.CrashShutdown
	CodeCannotConnectNowError                                  = pgcode.CannotConnectNow
	CodeDatabaseDroppedError                                   = pgcode.DatabaseDropped
	CodeSystemError                                            = pgcode.System
	CodeIoError                                                = pgcode.Io
	CodeUndefinedFileError                                     = pgcode.UndefinedFile
	CodeDuplicateFileError                                     = pgcode.DuplicateFile
	CodeConfigFileError                                        = pgcode.ConfigFile
	CodeLockFileExistsError                                    = pgcode.LockFileExists
	CodeFdwError                                               = pgcode.Fdw
	CodeFdwColumnNameNotFoundError                             = pgcode.FdwColumnNameNotFound
	CodeFdwDynamicParameterValueNeededError                    = pgcode.FdwDynamicParameterValueNeeded
	CodeFdwFunctionSequenceError                               = pgcode.FdwFunctionSequence
	CodeFdwInconsistentDescriptorInformationError              = pgcode.FdwInconsistentDescriptorInformation
	CodeFdwInvalidAttributeValueError                          = pgcode.FdwInvalidAttributeValue
	CodeFdwInvalidColumnNameError                              = pgcode.FdwInvalidColumnName
	CodeFdwInvalidColumnNumberError                            = pgcode.FdwInvalidColumnNumber
	CodeFdwInvalidDataTypeError                                = pgcode.FdwInvalidDataType
	CodeFdwInvalidDataTypeDescriptorsError                     = pgcode.FdwInvalidDataTypeDescriptors
	CodeFdwInvalidDescriptorFieldIdentifierError               = pgcode.FdwInvalidDescriptorFieldIdentifier
	CodeFdwInvalidHandleError                                  = pgcode.FdwInvalidHandle
	CodeFdwInvalidOptionIndexError                             = pgcode.FdwInvalidOptionIndex
	CodeFdwInvalidOptionNameError                              = pgcode.FdwInvalidOptionName
	CodeFdwInvalidStringLengthOrBufferLengthError              = pgcode.FdwInvalidStringLengthOrBufferLength
	CodeFdwInvalidStringFormatError                            = pgcode.FdwInvalidStringFormat
	CodeFdwInvalidUseOfNullPointerError                        = pgcode.FdwInvalidUseOfNullPointer
	CodeFdwTooManyHandlesError                                 = pgcode.FdwTooManyHandles
	CodeFdwOutOfMemoryError                                    = pgcode.FdwOutOfMemory
	CodeFdwNoSchemasError                                      = pgcode.FdwNoSchemas
	CodeFdwOptionNameNotFoundError                             = pgcode.FdwOptionNameNotFound
	CodeFdwReplyHandleError                                    = pgcode.FdwReplyHandle
	CodeFdwSchemaNotFoundError                                 = pgcode.FdwSchemaNotFound
	CodeFdwTableNotFoundError                                  = pgcode.FdwTableNotFound
	CodeFdwUnableToCreateExecutionError                        = pgcode.FdwUnableToCreateExecution
	CodeFdwUnableToCreateReplyError                            = pgcode.FdwUnableToCreateReply
	CodeFdwUnableToEstablishConnectionError                    = pgcode.FdwUnableToEstablishConnection
	CodePLpgSQLError                                           = pgcode.PLpgSQL
	CodeRaiseExceptionError                                    = pgcode.RaiseException
	CodeNoDataFoundError                                       = pgcode.NoDataFound
	CodeTooManyRowsError                                       = pgcode.TooManyRows
	CodeInternalError                                          = pgcode.Internal
	CodeDataCorruptedError                                     = pgcode.DataCorrupted
	CodeIndexCorruptedError                                    = pgcode.IndexCorrupted
	CodeUncategorizedError                                     = pgcode.Uncategorized
	CodeRangeUnavailable                                       = pgcode.RangeUnavailable
	CodeCCLRequired                                            = pgcode.CCLRequired
	CodeCCLValidLicenseRequired                                = pgcode.CCLValidLicenseRequired
	CodeTransactionCommittedWithSchemaChangeFailure            = pgcode.TransactionCommittedWithSchemaChangeFailure
)
