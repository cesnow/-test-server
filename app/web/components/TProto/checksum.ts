import {MessageType} from "@protobuf-ts/runtime";
import * as crc32 from "crc-32";
import * as allTypes from "./pb/gen";
import {TConstructor, TObject} from "@/components/TProto/manually_types";

const tProtoCrc = new Map<string, TConstructor>();
const crcTProto = new Map<TConstructor, MessageType<any>>();

export function initChecksum() {
  tProtoCrc.clear();
  crcTProto.clear();

  for (const type of Object.values(allTypes)) {
    const fullName = type.typeName;
    const checksum = crc32.str(fullName) >>> 0;
    const crc32Int: TConstructor = checksum > 0x7FFFFFFF
      ? checksum - 0x100000000
      : checksum;

    tProtoCrc.set(fullName, crc32Int);
    crcTProto.set(crc32Int, type);

    console.info(`crc[${crc32Int}] → ${fullName}`);
  }

  console.info(`[tProtoInit] Loaded types: ${tProtoCrc.size}`);
}

export function getTProtoChecksum(obj: TObject): TConstructor {
  const checksum = tProtoCrc.get(obj.typeName);
  return checksum !== undefined ? checksum : -1;
}

export function newTObject(checksum: TConstructor): MessageType<any> | null {
  return crcTProto.get(checksum) ?? null;
}

export function getTypeNameByChecksum(checksum: TConstructor): string | null {
  const type = crcTProto.get(checksum);
  return type?.typeName ?? null;
}
