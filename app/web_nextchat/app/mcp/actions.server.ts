"use server";
import * as actions from "./actions.base";

export const getAvailableClientsCount = actions.getAvailableClientsCount;
export const isMcpEnabled = actions.isMcpEnabled;
export const initializeMcpSystem = actions.initializeMcpSystem;
export const addMcpServer = actions.addMcpServer;
export const getClientsStatus = actions.getClientsStatus;
export const getClientTools = actions.getClientTools;
export const getMcpConfigFromFile = actions.getMcpConfigFromFile;
export const pauseMcpServer = actions.pauseMcpServer;
export const restartAllClients = actions.restartAllClients;
export const resumeMcpServer = actions.resumeMcpServer;
export const executeMcpAction = actions.executeMcpAction;
export const getAllTools = actions.getAllTools;
