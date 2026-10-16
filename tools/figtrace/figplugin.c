/* QEMU TCG plugin: trap figman PKEX crypto PLTs, dump guest buffers.
 * Host-side, no guest modification. Filters to figman via its .rodata signature.
 * Build: gcc -shared -fPIC -I. $(pkg-config --cflags glib-2.0) figplugin.c -o figplugin.so $(pkg-config --libs glib-2.0)
 */
#include <qemu-plugin.h>
#include <glib.h>
#include <stdio.h>
#include <string.h>
#include <inttypes.h>

QEMU_PLUGIN_EXPORT int qemu_plugin_version = QEMU_PLUGIN_VERSION;

static FILE *LOG;
static struct qemu_plugin_register *rsp_h;   /* stack pointer reg handle */

/* crypto PLT stubs for BOTH binaries (entry semantics: [esp]=ret, [esp+4]=arg0, ...).
 * bin: 0=figman (sig@0x807592d), 1=cmr (sig@0x809070a). */
enum { K_HKPDF, K_HMACCTOR, K_HMACUPD, K_HMACDIG, K_SIVENC, K_SIVDEC, K_EC };
static const struct { uint64_t addr; int kind; int bin; const char *name; } SITE[] = {
  /* figman */
  {0x80502e0, K_HKPDF,    0, "HKPDF"}, {0x80502d0, K_HMACCTOR, 0, "HMAC.ctor"},
  {0x804f120, K_HMACUPD,  0, "HMAC.upd"}, {0x80500d0, K_HMACDIG, 0, "HMAC.dig"},
  {0x804fc90, K_SIVENC,   0, "AESSIV.enc"}, {0x8050540, K_SIVDEC, 0, "AESSIV.dec"},
  /* cmr */
  {0x8050340, K_HKPDF,    1, "HKPDF"}, {0x8050330, K_HMACCTOR, 1, "HMAC.ctor"},
  {0x804f110, K_HMACUPD,  1, "HMAC.upd"}, {0x80500f0, K_HMACDIG, 1, "HMAC.dig"},
  {0x804fc30, K_SIVENC,   1, "AESSIV.enc"}, {0x8050570, K_SIVDEC, 1, "AESSIV.dec"},
  /* libucrypto.so function entries (base 0xf7edd000, ASLR off) — catches ALL callers */
  {0xf7eef106, K_HKPDF,    2, "HKPDF"}, {0xf7eedb8c, K_HMACCTOR, 2, "HMAC.ctor"},
  {0xf7eedc6e, K_HMACUPD,  2, "HMAC.upd"}, {0xf7eedcac, K_HMACDIG, 2, "HMAC.dig"},
  {0xf7ef6f6e, K_SIVENC,   2, "AESSIV.enc"}, {0xf7ef7078, K_SIVDEC, 2, "AESSIV.dec"},
  {0xf7ef3680, K_EC, 2, "WCurve.mul"},
  {0xf7ef2d60, K_EC, 2, "ECPoint.toBin"}, {0xf7ef2b70, K_EC, 2, "ECPoint.fromOct"},
  {0xf7efc554, K_EC, 2, "getECParams"}, {0xf7eedaf2, K_EC, 2, "SHA256.dig"},
};
#define NSITE (sizeof(SITE)/sizeof(SITE[0]))
static const uint64_t SIG_AT[2] = {0x807592dULL, 0x809070aULL};
static const char *BINNAME[3] = {"figman", "cmr", "libu"};
static const char SIG[] = "com.mikrotik.cmr";

static int rdmem(uint64_t addr, void *out, size_t len) {
  GByteArray *a = g_byte_array_new();
  int r = -1;
  if (qemu_plugin_read_memory_vaddr(addr, a, len) && a->len >= len) { memcpy(out, a->data, len); r = (int)len; }
  g_byte_array_free(a, TRUE);
  return r;
}
static uint32_t rd32(uint64_t a) { uint32_t v = 0; rdmem(a, &v, 4); return v; }

static uint32_t read_esp(void) {
  if (!rsp_h) return 0;
  GByteArray *b = g_byte_array_new();
  uint64_t rsp = 0;
  if (qemu_plugin_read_register(rsp_h, b) && b->len >= 4) memcpy(&rsp, b->data, b->len < 8 ? b->len : 8);
  g_byte_array_free(b, TRUE);
  return (uint32_t)rsp;
}

static void loghex(uint64_t addr, uint32_t n) {
  if (!addr || n == 0 || n > 4096) { fputs("\n", LOG); return; }
  unsigned char buf[4096];
  if (n > sizeof(buf)) n = sizeof(buf);
  if (rdmem(addr, buf, n) == (int)n) for (uint32_t i = 0; i < n; i++) fprintf(LOG, "%02x", buf[i]);
  fputc('\n', LOG);
}
/* dump vector<uchar> {begin,end,cap} at p -> label + hex */
static void logvec(uint32_t p, const char *tag) {
  fprintf(LOG, "%s", tag);
  if (!p) { fputs("<null>\n", LOG); return; }
  uint32_t b = rd32(p), e = rd32(p + 4);
  if (b && e >= b && (e - b) <= 4096) loghex(b, e - b); else fprintf(LOG, "<b=%08x e=%08x>\n", b, e);
}

static int sig_ok(int bin) {
  char s[sizeof(SIG) - 1];
  if (rdmem(SIG_AT[bin], s, sizeof(s)) != (int)sizeof(s)) return 0;
  return memcmp(s, SIG, sizeof(s)) == 0;
}

/* decode arg at [esp+off] as a vector<uchar> {begin,end,cap}; print if plausible */
static void tryvec(uint32_t esp, int off, const char *tag) {
  uint32_t p = rd32(esp + off);
  uint32_t b = rd32(p), e = rd32(p + 4);
  fprintf(LOG, "  [esp+%d]=%08x", off, p);
  if (b && e > b && (e - b) <= 4096) { fprintf(LOG, " vec[%u]=", e - b); loghex(b, e - b); }
  else fprintf(LOG, " (b=%08x e=%08x)\n", b, e);
}

static void on_exec(unsigned int vcpu, void *ud) {
  int si = (int)(intptr_t)ud;          /* SITE index */
  int kind = SITE[si].kind, bin = SITE[si].bin;
  if (bin < 2 && !sig_ok(bin)) return; /* exe PLTs: confirm binary; libu(bin2): log all */
  const char *proc = sig_ok(0) ? "figman" : (sig_ok(1) ? "cmr" : "?");
  uint32_t esp = read_esp();
  if (!esp) return;
  uint32_t caller = rd32(esp);
  uint32_t a0 = rd32(esp + 4), a1 = rd32(esp + 8), a2 = rd32(esp + 12), a3 = rd32(esp + 16), a4 = rd32(esp + 20), a5 = rd32(esp + 24);
  fprintf(LOG, "[%s/%s] %s caller=%08x args=%08x,%08x,%08x,%08x,%08x,%08x\n",
          BINNAME[bin], proc, SITE[si].name, caller, a0, a1, a2, a3, a4, a5);
  if (kind == K_HKPDF) { /* probe all arg slots as vectors to find salt/ikm/info */
    for (int off = 4; off <= 20; off += 4) tryvec(esp, off, "");
  } else if (kind == K_HMACCTOR) { /* this,Hash*,key,klen */
    fprintf(LOG, "  klen=%u key=", a3); loghex(a2, a3 <= 256 ? a3 : 0);
  } else if (kind == K_HMACUPD) { /* this,data,len */
    fprintf(LOG, "  ulen=%u data=", a2); loghex(a1, a2 <= 1024 ? a2 : 0);
  } else if (kind == K_HMACDIG) {
    fprintf(LOG, "  (out filled after return)\n");
  } else if (kind == K_EC) {
    /* WCurve.mul(sret,this,&k,&p): dump scalar+point for high-level (exe) callers */
    if (!strcmp(SITE[si].name,"WCurve.mul") && caller >= 0x8048000u && caller < 0x80a00000u) {
      fprintf(LOG, "  k.hdr="); loghex(a2, 16);
      fprintf(LOG, "  k.d0="); loghex(rd32(a2), 40);
      fprintf(LOG, "  k.d1="); loghex(rd32(a2+4), 40);
      fprintf(LOG, "  p.hdr="); loghex(a3, 16);
    } else if (!strcmp(SITE[si].name,"getECParams") || !strcmp(SITE[si].name,"ECPoint.fromOct")) {
      fprintf(LOG, "  a0deref="); loghex(rd32(a0), 40);
    }
  } else if (kind == K_SIVENC || kind == K_SIVDEC) { /* SIV: this,span,vec<span>,out */
    uint32_t sp = rd32(a1), ssz = rd32(a1 + 4);
    fprintf(LOG, "  %s[%u]=", kind == K_SIVENC ? "pt" : "ct", ssz); loghex(sp, ssz <= 1024 ? ssz : 0);
    fprintf(LOG, "  obj40="); loghex(a0, 64);
    uint32_t ab = rd32(a2), ae = rd32(a2 + 4);
    for (uint32_t i = ab; i && i + 8 <= ae && (ae - ab) <= 0x400; i += 8) { fprintf(LOG, "  AD="); loghex(rd32(i), rd32(i + 4)); }
  }
  fflush(LOG);
}

static void tb_trans(qemu_plugin_id_t id, struct qemu_plugin_tb *tb) {
  size_t n = qemu_plugin_tb_n_insns(tb);
  for (size_t i = 0; i < n; i++) {
    struct qemu_plugin_insn *insn = qemu_plugin_tb_get_insn(tb, i);
    uint64_t va = qemu_plugin_insn_vaddr(insn);
    for (size_t s = 0; s < NSITE; s++) {
      if (va == SITE[s].addr) {
        qemu_plugin_register_vcpu_insn_exec_cb(insn, on_exec, QEMU_PLUGIN_CB_R_REGS, (void *)(intptr_t)s);
      }
    }
  }
}

static void vcpu_init(qemu_plugin_id_t id, unsigned int vcpu) {
  if (rsp_h) return;
  GArray *regs = qemu_plugin_get_registers();
  if (!regs) return;
  for (guint i = 0; i < regs->len; i++) {
    qemu_plugin_reg_descriptor *d = &g_array_index(regs, qemu_plugin_reg_descriptor, i);
    if (d->name && (strcmp(d->name, "rsp") == 0 || strcmp(d->name, "esp") == 0)) { rsp_h = d->handle; }
  }
  fprintf(LOG, "=== vcpu%u init, rsp_h=%p ===\n", vcpu, (void *)rsp_h); fflush(LOG);
}

QEMU_PLUGIN_EXPORT int qemu_plugin_install(qemu_plugin_id_t id, const qemu_info_t *info, int argc, char **argv) {
  const char *path = argc > 0 ? argv[0] : "/tmp/figplugin.log";
  LOG = fopen(path, "a");
  if (!LOG) return -1;
  setvbuf(LOG, NULL, _IOLBF, 0);
  fprintf(LOG, "=== figplugin loaded ===\n"); fflush(LOG);
  qemu_plugin_register_vcpu_init_cb(id, vcpu_init);
  qemu_plugin_register_vcpu_tb_trans_cb(id, tb_trans);
  return 0;
}
