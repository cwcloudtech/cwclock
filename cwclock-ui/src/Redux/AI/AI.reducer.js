import {
  AiLOADING,
  AiERROR,
  AiProvidersSUCCESS,
  AiConnectionsSUCCESS,
  AiConversationsSUCCESS,
  AiConversationSUCCESS,
  AiConversationRESET,
  AiSENDING,
  AiSendDONE,
} from "./AI.types";

const initialState = {
  providers: [],
  connections: [],
  conversations: [],
  currentConversation: null,
  isLoading: false,
  isSending: false,
  isError: false,
};

export const AIReducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case AiLOADING:
      return { ...state, isLoading: true, isError: false };
    case AiERROR:
      return { ...state, isLoading: false, isSending: false, isError: true };
    case AiProvidersSUCCESS:
      return { ...state, providers: Array.isArray(payload) ? payload : [] };
    case AiConnectionsSUCCESS:
      return { ...state, connections: Array.isArray(payload) ? payload : [], isLoading: false };
    case AiConversationsSUCCESS:
      return { ...state, conversations: Array.isArray(payload) ? payload : [], isLoading: false };
    case AiConversationSUCCESS:
      return { ...state, currentConversation: payload, isLoading: false };
    case AiConversationRESET:
      return { ...state, currentConversation: null };
    case AiSENDING:
      return { ...state, isSending: true };
    case AiSendDONE:
      return { ...state, isSending: false };
    default:
      return state;
  }
};
