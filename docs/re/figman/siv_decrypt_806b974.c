
undefined4 FUN_0806b974(undefined4 param_1)

{
  int in_EAX;
  ostream *poVar1;
  uint in_ECX;
  undefined4 extraout_ECX;
  undefined4 uVar2;
  vector local_24 [4];
  uint local_20;
  int local_1c;
  int local_18;
  uchar local_14 [8];
  
  local_20 = in_ECX;
  if (DAT_0807c74c == 1) {
    if (((byte)runtimeDFLAGS & 2) == 0) goto LAB_0806b9e0;
  }
  else if (DAT_0807c74c + -1 < 0) goto LAB_0806b9e0;
  if (*(int *)(in_EAX + 0xc0) == *(int *)(in_EAX + 0xbc)) {
    poVar1 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
    poVar1 = (ostream *)
             operator<<(poVar1,"\"pkex/pkex.cpp\":1105: ECF\x1b[22m !encryptionKey.empty(): ");
    endl(poVar1);
    FUN_0806b4e6("!!");
  }
LAB_0806b9e0:
  if (local_20 < 0x10) {
    uVar2 = 0;
    poVar1 = (ostream *)
             operator<<(cout,"PKEXImpl::decrypt(span<const unsigned char>, vector<unsigned char>&)",
                        0x11,param_1);
    poVar1 = (ostream *)operator<<(poVar1,": wrapped data too short");
    endl(poVar1);
  }
  else {
    local_1c = *(int *)(in_EAX + 0xbc);
    local_18 = *(int *)(in_EAX + 0xc0) - local_1c;
    FUN_0805d046();
    local_14[0] = '\0';
    local_14[1] = '\0';
    local_14[2] = '\0';
    local_14[3] = '\0';
    local_14[4] = '\0';
    local_14[5] = '\0';
    local_14[6] = '\0';
    local_14[7] = '\0';
    uVar2 = AESSIV::decrypt((span *)&local_1c,local_24,local_14);
    FUN_0806b264();
    if ((char)uVar2 == '\0') {
      poVar1 = (ostream *)
               operator<<(cout,
                          "PKEXImpl::decrypt(span<const unsigned char>, vector<unsigned char>&)",
                          0x11,extraout_ECX);
      poVar1 = (ostream *)operator<<(poVar1,": AES-SIV decrypt failed");
      endl(poVar1);
    }
  }
  return uVar2;
}


