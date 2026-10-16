
void FUN_0806e820(void)

{
  Point *pPVar1;
  int in_EAX;
  ostream *poVar2;
  int iVar3;
  undefined4 extraout_ECX;
  undefined4 in_EDX;
  undefined4 extraout_EDX;
  undefined4 extraout_EDX_00;
  undefined1 *unaff_EBX;
  ECPoint *pEVar4;
  BigNum *pBVar5;
  undefined4 uVar6;
  undefined *puVar7;
  undefined1 *puVar8;
  void *local_194;
  int local_190;
  Point local_18c [16];
  BigNum local_17c [36];
  undefined **local_158 [36];
  undefined4 local_c8;
  undefined4 local_c4;
  undefined4 local_c0;
  
  if (DAT_0807c760 == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806e84a;
  }
  else if (-1 < DAT_0807c760 + -1) {
LAB_0806e84a:
    if (*(int *)(in_EAX + 0xa0) == 0) {
      unaff_EBX = cerr;
      poVar2 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      poVar2 = (ostream *)operator<<(poVar2,"\"pkex/pkex.cpp\":1052: ECF\x1b[22m kProt: ");
      endl(poVar2);
      FUN_0806b4e6("!!");
      in_EDX = extraout_EDX;
    }
  }
  if (DAT_0807c764 == 1) {
    if (((byte)runtimeDFLAGS & 2) == 0) goto LAB_0806e8dc;
  }
  else if (DAT_0807c764 + -1 < 0) goto LAB_0806e8dc;
  if (*(int *)(in_EAX + 0xb0) == 0) {
    puVar8 = unaff_EBX;
    poVar2 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
    puVar7 = &DAT_08077a4b;
    poVar2 = (ostream *)operator<<(poVar2,"\"pkex/pkex.cpp\":1053: ECF\x1b[22m ecPeerProt: ");
    endl(poVar2);
    FUN_0806b4e6("!!",puVar7,unaff_EBX,puVar8);
    in_EDX = extraout_EDX_00;
  }
LAB_0806e8dc:
  pPVar1 = *(Point **)(in_EAX + 0x58);
  WCurve::fromEC((ECPoint *)local_158);
  iVar3 = in_EAX + 0x98;
  pEVar4 = (ECPoint *)local_158;
  WCurve::mul(local_17c,pPVar1);
  WCurve::toEC(local_18c);
  FUN_0806e51c(iVar3,pEVar4,in_EDX);
  FUN_0806e51c();
  poVar2 = (ostream *)operator<<(cout,"PKEXImpl::reveal1(vector<unsigned char>&)",0x11,extraout_ECX)
  ;
  operator<<(poVar2,": J/L: ");
  poVar2 = (ostream *)FUN_0806b496();
  endl(poVar2);
  BigNum::bin((uint)&local_194);
  local_c4 = 0;
  local_c0 = 0;
  local_c8 = 0x807bcb0;
  HMAC::HMAC((HMAC *)local_158,(Hash *)&local_c8,local_194,local_190 - (int)local_194);
  iVar3 = in_EAX + 0x60;
  local_158[0] = &PTR_LAB_0807a158;
  pBVar5 = local_17c;
  BigNum::bin((uint)local_17c);
  FUN_0806a5ea();
  FUN_080521aa(local_17c,pBVar5,iVar3);
  uVar6 = *(undefined4 *)(in_EAX + 0xb0);
  pBVar5 = local_17c;
  BigNum::bin((uint)local_17c);
  FUN_0806a5ea();
  FUN_080521aa(local_17c,pBVar5,uVar6);
  iVar3 = *(int *)(in_EAX + 0xa0) + 4;
  pBVar5 = local_17c;
  BigNum::bin((uint)local_17c);
  FUN_0806a5ea();
  FUN_080521aa(local_17c,pBVar5,iVar3);
  pEVar4 = (ECPoint *)local_158;
  HMAC::digestLength();
  FUN_0805d046(pEVar4);
  HMAC::digest((uchar *)local_158);
  poVar2 = (ostream *)operator<<(cout,"PKEXImpl::reveal1(vector<unsigned char>&)",0x11);
  operator<<(poVar2,": u/v: ");
  FUN_0806b620();
  poVar2 = (ostream *)FUN_080514f9();
  endl(poVar2);
  string::freeptr();
  local_158[0] = &PTR_LAB_0807a158;
  HMAC::~HMAC((HMAC *)local_158);
  FUN_080521aa(&local_194);
  FUN_08052f88();
  return;
}


