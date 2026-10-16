
void FUN_0806e57e(void)

{
  vector *pvVar1;
  int in_EAX;
  ostream *poVar2;
  undefined4 uVar3;
  undefined4 in_ECX;
  undefined4 extraout_ECX;
  undefined4 extraout_ECX_00;
  undefined1 *unaff_ESI;
  undefined4 uVar4;
  undefined4 uVar5;
  undefined *puVar6;
  undefined1 *puVar7;
  vector *local_6c;
  undefined4 local_68;
  undefined4 local_64;
  undefined4 local_60;
  undefined4 local_5c;
  undefined1 local_58 [8];
  Point local_50 [16];
  Hash local_40 [48];
  
  if (DAT_0807c770 == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806e5a5;
  }
  else if (-1 < DAT_0807c770 + -1) {
LAB_0806e5a5:
    if (*(int *)(in_EAX + 0xb8) == 0) {
      poVar2 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      poVar2 = (ostream *)operator<<(poVar2,"\"pkex/pkex.cpp\":1030: ECF\x1b[22m code: ");
      endl(poVar2);
      FUN_0806b4e6("!!");
      in_ECX = extraout_ECX;
    }
  }
  if (DAT_0807c76c == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806e5f9;
  }
  else if (-1 < DAT_0807c76c + -1) {
LAB_0806e5f9:
    if (*(int *)(in_EAX + 0xa0) == 0) {
      unaff_ESI = cerr;
      uVar4 = in_ECX;
      poVar2 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      puVar6 = &DAT_0807798c;
      poVar2 = (ostream *)operator<<(poVar2,"\"pkex/pkex.cpp\":1031: ECF\x1b[22m kProt: ");
      endl(poVar2);
      FUN_0806b4e6("!!",puVar6,in_ECX,uVar4);
    }
  }
  if (DAT_0807c768 == 1) {
    if (((byte)runtimeDFLAGS & 2) == 0) goto LAB_0806e68b;
  }
  else if (DAT_0807c768 + -1 < 0) goto LAB_0806e68b;
  if (*(int *)(in_EAX + 0xb0) == 0) {
    puVar7 = unaff_ESI;
    poVar2 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
    puVar6 = &DAT_080779b3;
    poVar2 = (ostream *)operator<<(poVar2,"\"pkex/pkex.cpp\":1032: ECF\x1b[22m ecPeerProt: ");
    endl(poVar2);
    FUN_0806b4e6("!!",puVar6,unaff_ESI,puVar7);
  }
LAB_0806e68b:
  FUN_0806e53e(*(undefined4 *)(in_EAX + 0xb0));
  WCurve::toEC(local_50);
  FUN_0806e51c();
  poVar2 = (ostream *)
           operator<<(cout,"PKEXImpl::authDeriveZ(const BigNum&, const BigNum&)",0x15,
                      extraout_ECX_00);
  operator<<(poVar2,": K: ");
  poVar2 = (ostream *)FUN_0806b496();
  endl(poVar2);
  local_68 = 0;
  local_64 = 0;
  FUN_0806bcf6();
  FUN_0806bcf6();
  FUN_0806bcf6();
  uVar5 = 5;
  uVar3 = createHash((uint)&local_6c);
  uVar4 = uVar3;
  BigNum::bin((uint)local_58);
  local_60 = 0;
  local_5c = 0;
  HKPDF(local_40,local_6c,(vector *)&local_60,(vector *)&local_68,(uint)local_58);
  FUN_080522b8(uVar5);
  FUN_080521aa(local_40,uVar3,uVar4);
  FUN_080521aa(&local_60);
  FUN_080521aa(local_58);
  poVar2 = (ostream *)operator<<(cout,"PKEXImpl::authDeriveZ(const BigNum&, const BigNum&)",0x15);
  operator<<(poVar2,": z: ");
  FUN_0806b620();
  poVar2 = (ostream *)FUN_080514f9();
  endl(poVar2);
  string::freeptr();
  pvVar1 = local_6c;
  local_6c = (vector *)0x0;
  if (pvVar1 != (vector *)0x0) {
    (**(code **)(*(int *)pvVar1 + 8))(pvVar1);
  }
  FUN_080521aa((vector *)&local_68);
  FUN_08052f88();
  return;
}


