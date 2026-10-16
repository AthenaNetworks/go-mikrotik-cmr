
undefined4 FUN_0806b886(void)

{
  int in_EAX;
  ostream *poVar1;
  undefined4 uVar2;
  undefined4 extraout_ECX;
  int in_EDX;
  undefined *puVar3;
  int iVar4;
  int iVar5;
  int local_2c;
  int local_28;
  uchar local_24 [20];
  
  if (DAT_0807c750 == 1) {
    if (((byte)runtimeDFLAGS & 2) == 0) goto LAB_0806b8ee;
  }
  else if (DAT_0807c750 + -1 < 0) goto LAB_0806b8ee;
  iVar4 = *(int *)(in_EAX + 0xbc);
  if (*(int *)(in_EAX + 0xc0) == iVar4) {
    iVar5 = iVar4;
    poVar1 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
    puVar3 = &DAT_080771ab;
    poVar1 = (ostream *)
             operator<<(poVar1,"\"pkex/pkex.cpp\":1093: ECF\x1b[22m !encryptionKey.empty(): ");
    endl(poVar1);
    FUN_0806b4e6("!!",puVar3,iVar4,iVar5);
  }
LAB_0806b8ee:
  FUN_0805d046();
  local_2c = *(int *)(in_EAX + 0xbc);
  local_28 = *(int *)(in_EAX + 0xc0) - local_2c;
  local_24[0] = '\0';
  local_24[1] = '\0';
  local_24[2] = '\0';
  local_24[3] = '\0';
  local_24[4] = '\0';
  local_24[5] = '\0';
  local_24[6] = '\0';
  local_24[7] = '\0';
  uVar2 = AESSIV::encrypt((span *)&local_2c,(vector *)(in_EDX + 4),local_24);
  FUN_0806b264();
  if ((char)uVar2 == '\0') {
    poVar1 = (ostream *)
             operator<<(cout,"PKEXImpl::encrypt(m3::msg_view, vector<unsigned char>&)",0x11,
                        extraout_ECX);
    poVar1 = (ostream *)operator<<(poVar1,": AES-SIV encrypt failed");
    endl(poVar1);
  }
  return uVar2;
}


