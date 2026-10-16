/* Freestanding i386 ptrace tracer for figman crypto (no libc; raw int 0x80).
   Finds figman, breakpoints its crypto PLT stubs, logs args+buffers. */
typedef unsigned int u32; typedef int i32; typedef unsigned char u8;

static i32 sc(i32 n,i32 a,i32 b,i32 c,i32 d,i32 e){
  i32 r; __asm__ volatile("int $0x80":"=a"(r):"a"(n),"b"(a),"c"(b),"d"(c),"S"(d),"D"(e):"memory"); return r; }
#define SYS_read 3
#define SYS_write 4
#define SYS_open 5
#define SYS_close 6
#define SYS_exit 1
#define SYS_ptrace 26
#define SYS_wait4 114
#define SYS_getdents64 220
#define SYS_nanosleep 162
#define P_TRACEME 0
#define P_PEEK 1
#define P_POKE 4
#define P_CONT 7
#define P_STEP 9
#define P_GETREGS 12
#define P_SETREGS 13
#define P_ATTACH 16
#define P_DETACH 17

static i32 ptrace(i32 req,i32 pid,i32 addr,i32 data){ return sc(SYS_ptrace,req,pid,addr,data,0); }
static i32 wr(i32 fd,const void*b,i32 n){ return sc(SYS_write,fd,(i32)b,n,0,0); }
static i32 rd(i32 fd,void*b,i32 n){ return sc(SYS_read,fd,(i32)b,n,0,0); }
static i32 op(const char*p,i32 fl,i32 m){ return sc(SYS_open,(i32)p,fl,m,0,0); }
static void cl(i32 fd){ sc(SYS_close,fd,0,0,0,0); }
static void msleep(i32 ms){ i32 ts[2]={ms/1000,(ms%1000)*1000000}; sc(SYS_nanosleep,(i32)ts,0,0,0,0);}

static i32 slen(const char*s){ i32 n=0; while(s[n])n++; return n; }
static void us(char*d,const char*s){ while((*d++=*s++)); }
/* i386 user_regs_struct: ebx,ecx,edx,esi,edi,ebp,eax,ds,es,fs,gs,orig_eax,eip,cs,eflags,esp,ss */
struct regs{ u32 ebx,ecx,edx,esi,edi,ebp,eax,ds,es,fs,gs,oax,eip,cs,efl,esp,ss; };

static i32 LOG;
static void logs(const char*s){ wr(LOG,s,slen(s)); }
static const char hx[]="0123456789abcdef";
static void loghex(const u8*b,i32 n){ char t[2]; for(i32 i=0;i<n;i++){ t[0]=hx[b[i]>>4]; t[1]=hx[b[i]&15]; wr(LOG,t,2);} }
static void logu(u32 v){ char t[9]; for(i32 i=0;i<8;i++){ t[7-i]=hx[v&15]; v>>=4;} t[8]=0; logs(t); }

/* read n bytes from tracee addr into buf via PEEKTEXT (word at a time) */
static void peek(i32 pid,u32 addr,u8*buf,i32 n){
  for(i32 i=0;i<n;i+=4){ u32 w=0; ptrace(P_PEEK,pid,addr+i,(i32)&w); for(i32 j=0;j<4&&i+j<n;j++) buf[i+j]=(w>>(8*j))&0xff; } }
static u32 peek32(i32 pid,u32 addr){ u32 w=0; ptrace(P_PEEK,pid,addr,(i32)&w); return w; }

/* find pid whose /proc/<pid>/comm == name */
static i32 find_pid(const char*name){
  i32 fd=op("/proc",0,0); if(fd<0) return -1;
  char buf[4096]; i32 r;
  while((r=sc(SYS_getdents64,fd,(i32)buf,sizeof(buf),0,0))>0){
    i32 o=0;
    while(o<r){
      i32 ino_off=o; (void)ino_off;
      u32 reclen=*(unsigned short*)(buf+o+16);
      char*nm=buf+o+19;
      /* numeric? */
      if(nm[0]>='0'&&nm[0]<='9'){
        char path[64]; char*p=path; us(p,"/proc/"); p+=6; char*q=nm; while(*q)*p++=*q++; us(p,"/comm"); 
        i32 cf=op(path,0,0);
        if(cf>=0){ char c[64]; i32 cn=rd(cf,c,sizeof(c)-1); cl(cf); if(cn>0){ if(c[cn-1]=='\n')cn--; c[cn]=0;
          /* compare */ i32 eq=1; for(i32 k=0;name[k]||c[k];k++){ if(name[k]!=c[k]){eq=0;break;} } if(eq){ cl(fd);
            /* parse pid */ i32 pid=0; for(char*z=nm;*z;z++) pid=pid*10+(*z-'0'); return pid; } } }
      }
      o+=reclen;
    }
  }
  cl(fd); return -1;
}

/* breakpoint table: figman crypto PLTs */
#define NBP 6
static u32 BP[NBP]={0x80502e0,0x80502d0,0x804f120,0x80500d0,0x804fc90,0x8050540};
static const char* BPN[NBP]={"HKPDF","HMAC.ctor","HMAC.upd","HMAC.dig","SIV.enc","SIV.dec"};
static u32 ORIG[NBP];

static void set_bp(i32 pid,i32 i){ u32 w=peek32(pid,BP[i]); ORIG[i]=w; u32 nw=(w&0xffffff00)|0xcc; ptrace(P_POKE,pid,BP[i],(i32)nw); }
static void clr_bp(i32 pid,i32 i){ u32 w=peek32(pid,BP[i]); u32 nw=(w&0xffffff00)|(ORIG[i]&0xff); ptrace(P_POKE,pid,BP[i],(i32)nw); }

/* dump a vector<uchar> {begin,end,cap} at ptr */
static void dump_vec(i32 pid,u32 p,const char*tag){
  logs(tag); if(!p){logs("<null>\n");return;} u32 b=peek32(pid,p),e=peek32(pid,p+4);
  if(b&&e>=b&&(e-b)<=0x4000){ u8 tmp[1024]; i32 n=e-b; if(n>1024)n=1024; peek(pid,b,tmp,n); loghex(tmp,n);} logs("\n"); }

void _start(){
  LOG=op("/figtrace.log",01101,0644); /* O_CREAT|O_WRONLY|O_RDWR-ish */
  if(LOG<0) LOG=op("/ram/figtrace.log",01101,0644);
  if(LOG<0){ sc(SYS_exit,2,0,0,0,0); }
  logs("tracer start\n");
  i32 pid=-1; for(i32 tries=0;tries<200&&pid<0;tries++){ pid=find_pid("figman"); if(pid<0) msleep(50); }
  if(pid<0){ logs("figman not found\n"); sc(SYS_exit,3,0,0,0,0); }
  logs("figman pid="); logu((u32)pid); logs("\n");
  if(ptrace(P_ATTACH,pid,0,0)<0){ logs("attach failed\n"); sc(SYS_exit,4,0,0,0,0);}
  i32 st; sc(SYS_wait4,pid,(i32)&st,0,0,0);
  for(i32 i=0;i<NBP;i++) set_bp(pid,i);
  ptrace(P_CONT,pid,0,0);
  for(i32 loop=0;loop<100000;loop++){
    i32 rpid=sc(SYS_wait4,pid,(i32)&st,0,0,0); if(rpid<0) break;
    if((st&0x7f)==0){ logs("tracee exited\n"); break; }
    struct regs r; ptrace(P_GETREGS,pid,0,(i32)&r);
    u32 pc=r.eip-1; i32 hit=-1; for(i32 i=0;i<NBP;i++) if(pc==BP[i]){hit=i;break;}
    if(hit>=0){
      /* entry semantics: [esp]=ret, [esp+4]=arg0 ... */
      logs("HIT "); logs(BPN[hit]); logs(" caller="); logu(peek32(pid,r.esp)); logs("\n");
      u32 a0=peek32(pid,r.esp+4),a1=peek32(pid,r.esp+8),a2=peek32(pid,r.esp+12),a3=peek32(pid,r.esp+16),a4=peek32(pid,r.esp+20);
      if(hit==0){ /* HKPDF(out,A,B,C,len) */ dump_vec(pid,a1,"  A="); dump_vec(pid,a2,"  B="); dump_vec(pid,a3,"  C="); logs("  len="); logu(a4); logs("\n"); }
      else if(hit==1){ /* HMAC.ctor(this,Hash*,key,klen) */ logs("  klen="); logu(a3); logs(" key="); if(a2&&a3<=256){u8 t[256];peek(pid,a2,t,a3);loghex(t,a3);} logs("\n"); }
      else if(hit==2){ /* HMAC.upd(this,data,len) */ logs("  ulen="); logu(a2); logs(" data="); if(a1&&a2<=1024){u8 t[1024];peek(pid,a1,t,a2);loghex(t,a2);} logs("\n"); }
      else { /* others: just log first args */ logs("  a0="); logu(a0); logs(" a1="); logu(a1); logs(" a2="); logu(a2); logs(" a3="); logu(a3); logs("\n"); }
      /* step over bp */
      clr_bp(pid,hit); r.eip=pc; ptrace(P_SETREGS,pid,0,(i32)&r); ptrace(P_STEP,pid,0,0); sc(SYS_wait4,pid,(i32)&st,0,0,0); set_bp(pid,hit);
    }
    ptrace(P_CONT,pid,0,0);
  }
  for(i32 i=0;i<NBP;i++) clr_bp(pid,i);
  ptrace(P_DETACH,pid,0,0);
  logs("tracer done\n"); cl(LOG); sc(SYS_exit,0,0,0,0,0);
}
