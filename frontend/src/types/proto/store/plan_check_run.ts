// Code generated by protoc-gen-ts_proto. DO NOT EDIT.
// versions:
//   protoc-gen-ts_proto  v2.3.0
//   protoc               unknown
// source: store/plan_check_run.proto

/* eslint-disable */
import { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import Long from "long";
import { ChangedResources } from "./changelog";
import { Position } from "./common";

export const protobufPackage = "bytebase.store";

export interface PreUpdateBackupDetail {
  /**
   * The database for keeping the backup data.
   * Format: instances/{instance}/databases/{database}
   */
  database: string;
}

export interface PlanCheckRunConfig {
  sheetUid: number;
  changeDatabaseType: PlanCheckRunConfig_ChangeDatabaseType;
  instanceUid: number;
  databaseName: string;
  /** @deprecated */
  databaseGroupUid?: Long | undefined;
  ghostFlags: { [key: string]: string };
  /** If set, a backup of the modified data will be created automatically before any changes are applied. */
  preUpdateBackupDetail?: PreUpdateBackupDetail | undefined;
}

export enum PlanCheckRunConfig_ChangeDatabaseType {
  CHANGE_DATABASE_TYPE_UNSPECIFIED = "CHANGE_DATABASE_TYPE_UNSPECIFIED",
  DDL = "DDL",
  DML = "DML",
  SDL = "SDL",
  DDL_GHOST = "DDL_GHOST",
  SQL_EDITOR = "SQL_EDITOR",
  UNRECOGNIZED = "UNRECOGNIZED",
}

export function planCheckRunConfig_ChangeDatabaseTypeFromJSON(object: any): PlanCheckRunConfig_ChangeDatabaseType {
  switch (object) {
    case 0:
    case "CHANGE_DATABASE_TYPE_UNSPECIFIED":
      return PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED;
    case 1:
    case "DDL":
      return PlanCheckRunConfig_ChangeDatabaseType.DDL;
    case 2:
    case "DML":
      return PlanCheckRunConfig_ChangeDatabaseType.DML;
    case 3:
    case "SDL":
      return PlanCheckRunConfig_ChangeDatabaseType.SDL;
    case 4:
    case "DDL_GHOST":
      return PlanCheckRunConfig_ChangeDatabaseType.DDL_GHOST;
    case 5:
    case "SQL_EDITOR":
      return PlanCheckRunConfig_ChangeDatabaseType.SQL_EDITOR;
    case -1:
    case "UNRECOGNIZED":
    default:
      return PlanCheckRunConfig_ChangeDatabaseType.UNRECOGNIZED;
  }
}

export function planCheckRunConfig_ChangeDatabaseTypeToJSON(object: PlanCheckRunConfig_ChangeDatabaseType): string {
  switch (object) {
    case PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED:
      return "CHANGE_DATABASE_TYPE_UNSPECIFIED";
    case PlanCheckRunConfig_ChangeDatabaseType.DDL:
      return "DDL";
    case PlanCheckRunConfig_ChangeDatabaseType.DML:
      return "DML";
    case PlanCheckRunConfig_ChangeDatabaseType.SDL:
      return "SDL";
    case PlanCheckRunConfig_ChangeDatabaseType.DDL_GHOST:
      return "DDL_GHOST";
    case PlanCheckRunConfig_ChangeDatabaseType.SQL_EDITOR:
      return "SQL_EDITOR";
    case PlanCheckRunConfig_ChangeDatabaseType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export function planCheckRunConfig_ChangeDatabaseTypeToNumber(object: PlanCheckRunConfig_ChangeDatabaseType): number {
  switch (object) {
    case PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED:
      return 0;
    case PlanCheckRunConfig_ChangeDatabaseType.DDL:
      return 1;
    case PlanCheckRunConfig_ChangeDatabaseType.DML:
      return 2;
    case PlanCheckRunConfig_ChangeDatabaseType.SDL:
      return 3;
    case PlanCheckRunConfig_ChangeDatabaseType.DDL_GHOST:
      return 4;
    case PlanCheckRunConfig_ChangeDatabaseType.SQL_EDITOR:
      return 5;
    case PlanCheckRunConfig_ChangeDatabaseType.UNRECOGNIZED:
    default:
      return -1;
  }
}

export interface PlanCheckRunConfig_GhostFlagsEntry {
  key: string;
  value: string;
}

export interface PlanCheckRunResult {
  results: PlanCheckRunResult_Result[];
  error: string;
}

export interface PlanCheckRunResult_Result {
  status: PlanCheckRunResult_Result_Status;
  title: string;
  content: string;
  code: number;
  sqlSummaryReport?: PlanCheckRunResult_Result_SqlSummaryReport | undefined;
  sqlReviewReport?: PlanCheckRunResult_Result_SqlReviewReport | undefined;
}

export enum PlanCheckRunResult_Result_Status {
  STATUS_UNSPECIFIED = "STATUS_UNSPECIFIED",
  ERROR = "ERROR",
  WARNING = "WARNING",
  SUCCESS = "SUCCESS",
  UNRECOGNIZED = "UNRECOGNIZED",
}

export function planCheckRunResult_Result_StatusFromJSON(object: any): PlanCheckRunResult_Result_Status {
  switch (object) {
    case 0:
    case "STATUS_UNSPECIFIED":
      return PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED;
    case 1:
    case "ERROR":
      return PlanCheckRunResult_Result_Status.ERROR;
    case 2:
    case "WARNING":
      return PlanCheckRunResult_Result_Status.WARNING;
    case 3:
    case "SUCCESS":
      return PlanCheckRunResult_Result_Status.SUCCESS;
    case -1:
    case "UNRECOGNIZED":
    default:
      return PlanCheckRunResult_Result_Status.UNRECOGNIZED;
  }
}

export function planCheckRunResult_Result_StatusToJSON(object: PlanCheckRunResult_Result_Status): string {
  switch (object) {
    case PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED:
      return "STATUS_UNSPECIFIED";
    case PlanCheckRunResult_Result_Status.ERROR:
      return "ERROR";
    case PlanCheckRunResult_Result_Status.WARNING:
      return "WARNING";
    case PlanCheckRunResult_Result_Status.SUCCESS:
      return "SUCCESS";
    case PlanCheckRunResult_Result_Status.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}

export function planCheckRunResult_Result_StatusToNumber(object: PlanCheckRunResult_Result_Status): number {
  switch (object) {
    case PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED:
      return 0;
    case PlanCheckRunResult_Result_Status.ERROR:
      return 1;
    case PlanCheckRunResult_Result_Status.WARNING:
      return 2;
    case PlanCheckRunResult_Result_Status.SUCCESS:
      return 3;
    case PlanCheckRunResult_Result_Status.UNRECOGNIZED:
    default:
      return -1;
  }
}

export interface PlanCheckRunResult_Result_SqlSummaryReport {
  /** statement_types are the types of statements that are found in the sql. */
  statementTypes: string[];
  affectedRows: number;
  changedResources: ChangedResources | undefined;
}

export interface PlanCheckRunResult_Result_SqlReviewReport {
  line: number;
  column: number;
  /**
   * 1-based Position of the SQL statement.
   * To supersede `line` and `column` above.
   */
  startPosition: Position | undefined;
  endPosition: Position | undefined;
}

function createBasePreUpdateBackupDetail(): PreUpdateBackupDetail {
  return { database: "" };
}

export const PreUpdateBackupDetail: MessageFns<PreUpdateBackupDetail> = {
  encode(message: PreUpdateBackupDetail, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.database !== "") {
      writer.uint32(10).string(message.database);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PreUpdateBackupDetail {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePreUpdateBackupDetail();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 10) {
            break;
          }

          message.database = reader.string();
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PreUpdateBackupDetail {
    return { database: isSet(object.database) ? globalThis.String(object.database) : "" };
  },

  toJSON(message: PreUpdateBackupDetail): unknown {
    const obj: any = {};
    if (message.database !== "") {
      obj.database = message.database;
    }
    return obj;
  },

  create(base?: DeepPartial<PreUpdateBackupDetail>): PreUpdateBackupDetail {
    return PreUpdateBackupDetail.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<PreUpdateBackupDetail>): PreUpdateBackupDetail {
    const message = createBasePreUpdateBackupDetail();
    message.database = object.database ?? "";
    return message;
  },
};

function createBasePlanCheckRunConfig(): PlanCheckRunConfig {
  return {
    sheetUid: 0,
    changeDatabaseType: PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED,
    instanceUid: 0,
    databaseName: "",
    databaseGroupUid: undefined,
    ghostFlags: {},
    preUpdateBackupDetail: undefined,
  };
}

export const PlanCheckRunConfig: MessageFns<PlanCheckRunConfig> = {
  encode(message: PlanCheckRunConfig, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.sheetUid !== 0) {
      writer.uint32(8).int32(message.sheetUid);
    }
    if (message.changeDatabaseType !== PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED) {
      writer.uint32(16).int32(planCheckRunConfig_ChangeDatabaseTypeToNumber(message.changeDatabaseType));
    }
    if (message.instanceUid !== 0) {
      writer.uint32(24).int32(message.instanceUid);
    }
    if (message.databaseName !== "") {
      writer.uint32(34).string(message.databaseName);
    }
    if (message.databaseGroupUid !== undefined) {
      writer.uint32(40).int64(message.databaseGroupUid.toString());
    }
    Object.entries(message.ghostFlags).forEach(([key, value]) => {
      PlanCheckRunConfig_GhostFlagsEntry.encode({ key: key as any, value }, writer.uint32(50).fork()).join();
    });
    if (message.preUpdateBackupDetail !== undefined) {
      PreUpdateBackupDetail.encode(message.preUpdateBackupDetail, writer.uint32(58).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunConfig {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunConfig();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 8) {
            break;
          }

          message.sheetUid = reader.int32();
          continue;
        }
        case 2: {
          if (tag !== 16) {
            break;
          }

          message.changeDatabaseType = planCheckRunConfig_ChangeDatabaseTypeFromJSON(reader.int32());
          continue;
        }
        case 3: {
          if (tag !== 24) {
            break;
          }

          message.instanceUid = reader.int32();
          continue;
        }
        case 4: {
          if (tag !== 34) {
            break;
          }

          message.databaseName = reader.string();
          continue;
        }
        case 5: {
          if (tag !== 40) {
            break;
          }

          message.databaseGroupUid = Long.fromString(reader.int64().toString());
          continue;
        }
        case 6: {
          if (tag !== 50) {
            break;
          }

          const entry6 = PlanCheckRunConfig_GhostFlagsEntry.decode(reader, reader.uint32());
          if (entry6.value !== undefined) {
            message.ghostFlags[entry6.key] = entry6.value;
          }
          continue;
        }
        case 7: {
          if (tag !== 58) {
            break;
          }

          message.preUpdateBackupDetail = PreUpdateBackupDetail.decode(reader, reader.uint32());
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunConfig {
    return {
      sheetUid: isSet(object.sheetUid) ? globalThis.Number(object.sheetUid) : 0,
      changeDatabaseType: isSet(object.changeDatabaseType)
        ? planCheckRunConfig_ChangeDatabaseTypeFromJSON(object.changeDatabaseType)
        : PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED,
      instanceUid: isSet(object.instanceUid) ? globalThis.Number(object.instanceUid) : 0,
      databaseName: isSet(object.databaseName) ? globalThis.String(object.databaseName) : "",
      databaseGroupUid: isSet(object.databaseGroupUid) ? Long.fromValue(object.databaseGroupUid) : undefined,
      ghostFlags: isObject(object.ghostFlags)
        ? Object.entries(object.ghostFlags).reduce<{ [key: string]: string }>((acc, [key, value]) => {
          acc[key] = String(value);
          return acc;
        }, {})
        : {},
      preUpdateBackupDetail: isSet(object.preUpdateBackupDetail)
        ? PreUpdateBackupDetail.fromJSON(object.preUpdateBackupDetail)
        : undefined,
    };
  },

  toJSON(message: PlanCheckRunConfig): unknown {
    const obj: any = {};
    if (message.sheetUid !== 0) {
      obj.sheetUid = Math.round(message.sheetUid);
    }
    if (message.changeDatabaseType !== PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED) {
      obj.changeDatabaseType = planCheckRunConfig_ChangeDatabaseTypeToJSON(message.changeDatabaseType);
    }
    if (message.instanceUid !== 0) {
      obj.instanceUid = Math.round(message.instanceUid);
    }
    if (message.databaseName !== "") {
      obj.databaseName = message.databaseName;
    }
    if (message.databaseGroupUid !== undefined) {
      obj.databaseGroupUid = (message.databaseGroupUid || Long.ZERO).toString();
    }
    if (message.ghostFlags) {
      const entries = Object.entries(message.ghostFlags);
      if (entries.length > 0) {
        obj.ghostFlags = {};
        entries.forEach(([k, v]) => {
          obj.ghostFlags[k] = v;
        });
      }
    }
    if (message.preUpdateBackupDetail !== undefined) {
      obj.preUpdateBackupDetail = PreUpdateBackupDetail.toJSON(message.preUpdateBackupDetail);
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunConfig>): PlanCheckRunConfig {
    return PlanCheckRunConfig.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<PlanCheckRunConfig>): PlanCheckRunConfig {
    const message = createBasePlanCheckRunConfig();
    message.sheetUid = object.sheetUid ?? 0;
    message.changeDatabaseType = object.changeDatabaseType ??
      PlanCheckRunConfig_ChangeDatabaseType.CHANGE_DATABASE_TYPE_UNSPECIFIED;
    message.instanceUid = object.instanceUid ?? 0;
    message.databaseName = object.databaseName ?? "";
    message.databaseGroupUid = (object.databaseGroupUid !== undefined && object.databaseGroupUid !== null)
      ? Long.fromValue(object.databaseGroupUid)
      : undefined;
    message.ghostFlags = Object.entries(object.ghostFlags ?? {}).reduce<{ [key: string]: string }>(
      (acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = globalThis.String(value);
        }
        return acc;
      },
      {},
    );
    message.preUpdateBackupDetail =
      (object.preUpdateBackupDetail !== undefined && object.preUpdateBackupDetail !== null)
        ? PreUpdateBackupDetail.fromPartial(object.preUpdateBackupDetail)
        : undefined;
    return message;
  },
};

function createBasePlanCheckRunConfig_GhostFlagsEntry(): PlanCheckRunConfig_GhostFlagsEntry {
  return { key: "", value: "" };
}

export const PlanCheckRunConfig_GhostFlagsEntry: MessageFns<PlanCheckRunConfig_GhostFlagsEntry> = {
  encode(message: PlanCheckRunConfig_GhostFlagsEntry, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.key !== "") {
      writer.uint32(10).string(message.key);
    }
    if (message.value !== "") {
      writer.uint32(18).string(message.value);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunConfig_GhostFlagsEntry {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunConfig_GhostFlagsEntry();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 10) {
            break;
          }

          message.key = reader.string();
          continue;
        }
        case 2: {
          if (tag !== 18) {
            break;
          }

          message.value = reader.string();
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunConfig_GhostFlagsEntry {
    return {
      key: isSet(object.key) ? globalThis.String(object.key) : "",
      value: isSet(object.value) ? globalThis.String(object.value) : "",
    };
  },

  toJSON(message: PlanCheckRunConfig_GhostFlagsEntry): unknown {
    const obj: any = {};
    if (message.key !== "") {
      obj.key = message.key;
    }
    if (message.value !== "") {
      obj.value = message.value;
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunConfig_GhostFlagsEntry>): PlanCheckRunConfig_GhostFlagsEntry {
    return PlanCheckRunConfig_GhostFlagsEntry.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<PlanCheckRunConfig_GhostFlagsEntry>): PlanCheckRunConfig_GhostFlagsEntry {
    const message = createBasePlanCheckRunConfig_GhostFlagsEntry();
    message.key = object.key ?? "";
    message.value = object.value ?? "";
    return message;
  },
};

function createBasePlanCheckRunResult(): PlanCheckRunResult {
  return { results: [], error: "" };
}

export const PlanCheckRunResult: MessageFns<PlanCheckRunResult> = {
  encode(message: PlanCheckRunResult, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    for (const v of message.results) {
      PlanCheckRunResult_Result.encode(v!, writer.uint32(10).fork()).join();
    }
    if (message.error !== "") {
      writer.uint32(18).string(message.error);
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunResult {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunResult();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 10) {
            break;
          }

          message.results.push(PlanCheckRunResult_Result.decode(reader, reader.uint32()));
          continue;
        }
        case 2: {
          if (tag !== 18) {
            break;
          }

          message.error = reader.string();
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunResult {
    return {
      results: globalThis.Array.isArray(object?.results)
        ? object.results.map((e: any) => PlanCheckRunResult_Result.fromJSON(e))
        : [],
      error: isSet(object.error) ? globalThis.String(object.error) : "",
    };
  },

  toJSON(message: PlanCheckRunResult): unknown {
    const obj: any = {};
    if (message.results?.length) {
      obj.results = message.results.map((e) => PlanCheckRunResult_Result.toJSON(e));
    }
    if (message.error !== "") {
      obj.error = message.error;
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunResult>): PlanCheckRunResult {
    return PlanCheckRunResult.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<PlanCheckRunResult>): PlanCheckRunResult {
    const message = createBasePlanCheckRunResult();
    message.results = object.results?.map((e) => PlanCheckRunResult_Result.fromPartial(e)) || [];
    message.error = object.error ?? "";
    return message;
  },
};

function createBasePlanCheckRunResult_Result(): PlanCheckRunResult_Result {
  return {
    status: PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED,
    title: "",
    content: "",
    code: 0,
    sqlSummaryReport: undefined,
    sqlReviewReport: undefined,
  };
}

export const PlanCheckRunResult_Result: MessageFns<PlanCheckRunResult_Result> = {
  encode(message: PlanCheckRunResult_Result, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.status !== PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED) {
      writer.uint32(8).int32(planCheckRunResult_Result_StatusToNumber(message.status));
    }
    if (message.title !== "") {
      writer.uint32(18).string(message.title);
    }
    if (message.content !== "") {
      writer.uint32(26).string(message.content);
    }
    if (message.code !== 0) {
      writer.uint32(32).int32(message.code);
    }
    if (message.sqlSummaryReport !== undefined) {
      PlanCheckRunResult_Result_SqlSummaryReport.encode(message.sqlSummaryReport, writer.uint32(42).fork()).join();
    }
    if (message.sqlReviewReport !== undefined) {
      PlanCheckRunResult_Result_SqlReviewReport.encode(message.sqlReviewReport, writer.uint32(50).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunResult_Result {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunResult_Result();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 8) {
            break;
          }

          message.status = planCheckRunResult_Result_StatusFromJSON(reader.int32());
          continue;
        }
        case 2: {
          if (tag !== 18) {
            break;
          }

          message.title = reader.string();
          continue;
        }
        case 3: {
          if (tag !== 26) {
            break;
          }

          message.content = reader.string();
          continue;
        }
        case 4: {
          if (tag !== 32) {
            break;
          }

          message.code = reader.int32();
          continue;
        }
        case 5: {
          if (tag !== 42) {
            break;
          }

          message.sqlSummaryReport = PlanCheckRunResult_Result_SqlSummaryReport.decode(reader, reader.uint32());
          continue;
        }
        case 6: {
          if (tag !== 50) {
            break;
          }

          message.sqlReviewReport = PlanCheckRunResult_Result_SqlReviewReport.decode(reader, reader.uint32());
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunResult_Result {
    return {
      status: isSet(object.status)
        ? planCheckRunResult_Result_StatusFromJSON(object.status)
        : PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED,
      title: isSet(object.title) ? globalThis.String(object.title) : "",
      content: isSet(object.content) ? globalThis.String(object.content) : "",
      code: isSet(object.code) ? globalThis.Number(object.code) : 0,
      sqlSummaryReport: isSet(object.sqlSummaryReport)
        ? PlanCheckRunResult_Result_SqlSummaryReport.fromJSON(object.sqlSummaryReport)
        : undefined,
      sqlReviewReport: isSet(object.sqlReviewReport)
        ? PlanCheckRunResult_Result_SqlReviewReport.fromJSON(object.sqlReviewReport)
        : undefined,
    };
  },

  toJSON(message: PlanCheckRunResult_Result): unknown {
    const obj: any = {};
    if (message.status !== PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED) {
      obj.status = planCheckRunResult_Result_StatusToJSON(message.status);
    }
    if (message.title !== "") {
      obj.title = message.title;
    }
    if (message.content !== "") {
      obj.content = message.content;
    }
    if (message.code !== 0) {
      obj.code = Math.round(message.code);
    }
    if (message.sqlSummaryReport !== undefined) {
      obj.sqlSummaryReport = PlanCheckRunResult_Result_SqlSummaryReport.toJSON(message.sqlSummaryReport);
    }
    if (message.sqlReviewReport !== undefined) {
      obj.sqlReviewReport = PlanCheckRunResult_Result_SqlReviewReport.toJSON(message.sqlReviewReport);
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunResult_Result>): PlanCheckRunResult_Result {
    return PlanCheckRunResult_Result.fromPartial(base ?? {});
  },
  fromPartial(object: DeepPartial<PlanCheckRunResult_Result>): PlanCheckRunResult_Result {
    const message = createBasePlanCheckRunResult_Result();
    message.status = object.status ?? PlanCheckRunResult_Result_Status.STATUS_UNSPECIFIED;
    message.title = object.title ?? "";
    message.content = object.content ?? "";
    message.code = object.code ?? 0;
    message.sqlSummaryReport = (object.sqlSummaryReport !== undefined && object.sqlSummaryReport !== null)
      ? PlanCheckRunResult_Result_SqlSummaryReport.fromPartial(object.sqlSummaryReport)
      : undefined;
    message.sqlReviewReport = (object.sqlReviewReport !== undefined && object.sqlReviewReport !== null)
      ? PlanCheckRunResult_Result_SqlReviewReport.fromPartial(object.sqlReviewReport)
      : undefined;
    return message;
  },
};

function createBasePlanCheckRunResult_Result_SqlSummaryReport(): PlanCheckRunResult_Result_SqlSummaryReport {
  return { statementTypes: [], affectedRows: 0, changedResources: undefined };
}

export const PlanCheckRunResult_Result_SqlSummaryReport: MessageFns<PlanCheckRunResult_Result_SqlSummaryReport> = {
  encode(message: PlanCheckRunResult_Result_SqlSummaryReport, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    for (const v of message.statementTypes) {
      writer.uint32(18).string(v!);
    }
    if (message.affectedRows !== 0) {
      writer.uint32(24).int32(message.affectedRows);
    }
    if (message.changedResources !== undefined) {
      ChangedResources.encode(message.changedResources, writer.uint32(34).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunResult_Result_SqlSummaryReport {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunResult_Result_SqlSummaryReport();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2: {
          if (tag !== 18) {
            break;
          }

          message.statementTypes.push(reader.string());
          continue;
        }
        case 3: {
          if (tag !== 24) {
            break;
          }

          message.affectedRows = reader.int32();
          continue;
        }
        case 4: {
          if (tag !== 34) {
            break;
          }

          message.changedResources = ChangedResources.decode(reader, reader.uint32());
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunResult_Result_SqlSummaryReport {
    return {
      statementTypes: globalThis.Array.isArray(object?.statementTypes)
        ? object.statementTypes.map((e: any) => globalThis.String(e))
        : [],
      affectedRows: isSet(object.affectedRows) ? globalThis.Number(object.affectedRows) : 0,
      changedResources: isSet(object.changedResources) ? ChangedResources.fromJSON(object.changedResources) : undefined,
    };
  },

  toJSON(message: PlanCheckRunResult_Result_SqlSummaryReport): unknown {
    const obj: any = {};
    if (message.statementTypes?.length) {
      obj.statementTypes = message.statementTypes;
    }
    if (message.affectedRows !== 0) {
      obj.affectedRows = Math.round(message.affectedRows);
    }
    if (message.changedResources !== undefined) {
      obj.changedResources = ChangedResources.toJSON(message.changedResources);
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunResult_Result_SqlSummaryReport>): PlanCheckRunResult_Result_SqlSummaryReport {
    return PlanCheckRunResult_Result_SqlSummaryReport.fromPartial(base ?? {});
  },
  fromPartial(
    object: DeepPartial<PlanCheckRunResult_Result_SqlSummaryReport>,
  ): PlanCheckRunResult_Result_SqlSummaryReport {
    const message = createBasePlanCheckRunResult_Result_SqlSummaryReport();
    message.statementTypes = object.statementTypes?.map((e) => e) || [];
    message.affectedRows = object.affectedRows ?? 0;
    message.changedResources = (object.changedResources !== undefined && object.changedResources !== null)
      ? ChangedResources.fromPartial(object.changedResources)
      : undefined;
    return message;
  },
};

function createBasePlanCheckRunResult_Result_SqlReviewReport(): PlanCheckRunResult_Result_SqlReviewReport {
  return { line: 0, column: 0, startPosition: undefined, endPosition: undefined };
}

export const PlanCheckRunResult_Result_SqlReviewReport: MessageFns<PlanCheckRunResult_Result_SqlReviewReport> = {
  encode(message: PlanCheckRunResult_Result_SqlReviewReport, writer: BinaryWriter = new BinaryWriter()): BinaryWriter {
    if (message.line !== 0) {
      writer.uint32(8).int32(message.line);
    }
    if (message.column !== 0) {
      writer.uint32(16).int32(message.column);
    }
    if (message.startPosition !== undefined) {
      Position.encode(message.startPosition, writer.uint32(66).fork()).join();
    }
    if (message.endPosition !== undefined) {
      Position.encode(message.endPosition, writer.uint32(74).fork()).join();
    }
    return writer;
  },

  decode(input: BinaryReader | Uint8Array, length?: number): PlanCheckRunResult_Result_SqlReviewReport {
    const reader = input instanceof BinaryReader ? input : new BinaryReader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBasePlanCheckRunResult_Result_SqlReviewReport();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1: {
          if (tag !== 8) {
            break;
          }

          message.line = reader.int32();
          continue;
        }
        case 2: {
          if (tag !== 16) {
            break;
          }

          message.column = reader.int32();
          continue;
        }
        case 8: {
          if (tag !== 66) {
            break;
          }

          message.startPosition = Position.decode(reader, reader.uint32());
          continue;
        }
        case 9: {
          if (tag !== 74) {
            break;
          }

          message.endPosition = Position.decode(reader, reader.uint32());
          continue;
        }
      }
      if ((tag & 7) === 4 || tag === 0) {
        break;
      }
      reader.skip(tag & 7);
    }
    return message;
  },

  fromJSON(object: any): PlanCheckRunResult_Result_SqlReviewReport {
    return {
      line: isSet(object.line) ? globalThis.Number(object.line) : 0,
      column: isSet(object.column) ? globalThis.Number(object.column) : 0,
      startPosition: isSet(object.startPosition) ? Position.fromJSON(object.startPosition) : undefined,
      endPosition: isSet(object.endPosition) ? Position.fromJSON(object.endPosition) : undefined,
    };
  },

  toJSON(message: PlanCheckRunResult_Result_SqlReviewReport): unknown {
    const obj: any = {};
    if (message.line !== 0) {
      obj.line = Math.round(message.line);
    }
    if (message.column !== 0) {
      obj.column = Math.round(message.column);
    }
    if (message.startPosition !== undefined) {
      obj.startPosition = Position.toJSON(message.startPosition);
    }
    if (message.endPosition !== undefined) {
      obj.endPosition = Position.toJSON(message.endPosition);
    }
    return obj;
  },

  create(base?: DeepPartial<PlanCheckRunResult_Result_SqlReviewReport>): PlanCheckRunResult_Result_SqlReviewReport {
    return PlanCheckRunResult_Result_SqlReviewReport.fromPartial(base ?? {});
  },
  fromPartial(
    object: DeepPartial<PlanCheckRunResult_Result_SqlReviewReport>,
  ): PlanCheckRunResult_Result_SqlReviewReport {
    const message = createBasePlanCheckRunResult_Result_SqlReviewReport();
    message.line = object.line ?? 0;
    message.column = object.column ?? 0;
    message.startPosition = (object.startPosition !== undefined && object.startPosition !== null)
      ? Position.fromPartial(object.startPosition)
      : undefined;
    message.endPosition = (object.endPosition !== undefined && object.endPosition !== null)
      ? Position.fromPartial(object.endPosition)
      : undefined;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | Long : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

function isObject(value: any): boolean {
  return typeof value === "object" && value !== null;
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}

export interface MessageFns<T> {
  encode(message: T, writer?: BinaryWriter): BinaryWriter;
  decode(input: BinaryReader | Uint8Array, length?: number): T;
  fromJSON(object: any): T;
  toJSON(message: T): unknown;
  create(base?: DeepPartial<T>): T;
  fromPartial(object: DeepPartial<T>): T;
}
