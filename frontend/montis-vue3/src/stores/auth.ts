import { reactive } from 'vue';
import { authApi } from '@/api';
import { tokenStorage, setUnauthorizedHandler } from '@/api/client';
import type { User } from '@/types';
export const auth = reactive<{user:User|null; loading:boolean}>({user:null,loading:false});
setUnauthorizedHandler(()=>{auth.user=null;tokenStorage.clear()});
export async function restoreSession(){if(!tokenStorage.get())return;try{auth.user=await authApi.me()}catch{auth.user=null;tokenStorage.clear()}}
export async function signIn(email:string,password:string){const data=await authApi.login({email,password});tokenStorage.set(data.token);auth.user=data.user}
export async function signUp(name:string,phone:string,email:string,password:string){const data=await authApi.register({name,phone,email,password});tokenStorage.set(data.token);auth.user=data.user}
export function signOut(){tokenStorage.clear();auth.user=null;authApi.logout().catch(()=>{})}
