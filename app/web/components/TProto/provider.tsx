'use client';

import React, {Context, createContext, useContext} from "react";
import {useVLProto, VLProtoProps} from "@/components/TProto/core-network";

const TProtoContext: Context<VLProtoProps> = createContext<VLProtoProps>({} as VLProtoProps);

export const TProtoProvider = ({children}: { children: React.ReactNode }) => {

  const vlProto: VLProtoProps = useVLProto();

  return (
    <TProtoContext.Provider value={vlProto}>
      {children}
    </TProtoContext.Provider>
  )
}

export const useTProtoCtx = () => useContext(TProtoContext);
