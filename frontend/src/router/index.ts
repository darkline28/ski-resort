import {createRouter,createWebHistory} from 'vue-router';
import {auth,restoreSession} from '@/stores/auth';
const router=createRouter({history:createWebHistory(),routes:[
{path:'/',component:()=>import('@/pages/HomePage.vue')},
{path:'/lessons',component:()=>import('@/pages/CatalogPage.vue'),props:{kind:'lessons'}},
{path:'/rental',redirect:'/equipment'},
{path:'/equipment',component:()=>import('@/pages/CatalogPage.vue'),props:{kind:'equipment'}},
{path:'/instructors',component:()=>import('@/pages/CatalogPage.vue'),props:{kind:'instructors'}},
{path:'/login',component:()=>import('@/pages/AuthPage.vue'),props:{mode:'login'}},
{path:'/register',component:()=>import('@/pages/AuthPage.vue'),props:{mode:'register'}},
{path:'/booking',component:()=>import('@/pages/BookingPage.vue')},
{path:'/account',component:()=>import('@/pages/AccountPage.vue'),meta:{role:'client'}},
{path:'/instructor',component:()=>import('@/pages/InstructorPage.vue'),meta:{role:'instructor'}},
{path:'/admin',component:()=>import('@/pages/AdminPage.vue'),meta:{role:'admin'}},
{path:'/:pathMatch(.*)*',component:()=>import('@/pages/NotFoundPage.vue')}
]});
let initialized=false;
router.beforeEach(async(to)=>{if(!initialized){initialized=true;await restoreSession()}const role=to.meta.role;if(role&&!auth.user)return {path:'/login',query:{next:to.fullPath}};if(role&&auth.user?.role!==role)return '/';});
export default router;
