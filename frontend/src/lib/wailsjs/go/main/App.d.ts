// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {match} from '../models';

export function AddMatchesToQuickfixList():Promise<void>;

export function GetAppState():Promise<main.AppState>;

export function GetNextPage():Promise<main.SearchResult>;

export function GetPrevPage():Promise<main.SearchResult>;

export function GetReplacementText(arg1:string,arg2:string,arg3:string,arg4:boolean):Promise<string>;

export function GetRoute():Promise<string>;

export function OpenMatch(arg1:string,arg2:number,arg3:number):Promise<void>;

export function Replace(arg1:match.Match,arg2:string,arg3:string,arg4:boolean):Promise<void>;

export function ReplaceAll(arg1:string,arg2:string,arg3:string,arg4:string,arg5:string,arg6:boolean,arg7:boolean,arg8:boolean,arg9:boolean):Promise<void>;

export function Search(arg1:string,arg2:string,arg3:string,arg4:string,arg5:string,arg6:boolean,arg7:boolean,arg8:boolean,arg9:boolean):Promise<main.SearchResult>;

export function SendKey(arg1:string,arg2:boolean,arg3:boolean,arg4:boolean):Promise<void>;

export function Undo():Promise<void>;
