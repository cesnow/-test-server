'use client';

import React, {Context, createContext, useContext} from "react";
import {TProtoCore, TProtoProps} from "@/components/TProto/core-network";

const TProtoContext: Context<TProtoProps> = createContext<TProtoProps>({} as TProtoProps);

export const TProtoProvider = ({children}: { children: React.ReactNode }) => {

  const vlProto: TProtoProps = TProtoCore();

  return (
    <TProtoContext.Provider value={vlProto}>
      {children}
    </TProtoContext.Provider>
  )
}

export const useTProto = () => useContext(TProtoContext);
