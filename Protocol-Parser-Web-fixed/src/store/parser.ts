import axios from "axios";
import { create } from "zustand";

export interface ParseField { index:number; name:string; offset:number; length:number; raw:string; value:string; description:string }
export interface ParseResult { protocol:string; messageId:string; messageName:string; length:number; raw:string; fields:ParseField[]; data:unknown }

interface ParserStore {
    protocol: string;
    hex: string;
    result: ParseResult | null;
    loading: boolean;
    error: string | null;
    historyVersion: number;
    setProtocol: (protocol:string) => void;
    setHex: (hex:string) => void;
    parse: () => Promise<void>;
    clear: () => void;
    bumpHistoryVersion: () => void;
}

export const useParserStore=create<ParserStore>((set,get)=>({
    protocol:"2929", hex:"", result:null, loading:false, error:null, historyVersion:0,
    setProtocol:(protocol)=>set({protocol,result:null,error:null}),
    setHex:(hex)=>set({hex,error:null}),
    parse:async()=>{
        const {protocol,hex}=get(); set({loading:true,error:null});
        try { const response=await axios.post("/api/parser/analyze",{protocol,hex}); set((state)=>({result:response.data.data,loading:false,historyVersion:state.historyVersion+1})); }
        catch (error) { const message=axios.isAxiosError(error)?(error.response?.data?.error??error.message):"解析失败";set({result:null,loading:false,error:String(message)});throw error; }
    },
    clear:()=>set({hex:"",result:null,error:null}),
    bumpHistoryVersion:()=>set((state)=>({historyVersion:state.historyVersion+1}))
}));
