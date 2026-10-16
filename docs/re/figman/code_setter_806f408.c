
string * FUN_0806f408(void)

{
  int iVar1;
  int iVar2;
  int iVar3;
  char cVar4;
  string *in_EAX;
  ostream *poVar5;
  undefined4 *puVar6;
  int *piVar7;
  int iVar8;
  int *in_EDX;
  int *piVar9;
  int *piVar10;
  byte bVar11;
  undefined *puVar12;
  char *pcVar13;
  int *local_78;
  undefined4 local_24;
  undefined4 local_20;
  
  bVar11 = 0;
  poVar5 = (ostream *)operator<<(cout,"PKEX::start()",0xb);
  endl(poVar5);
  FUN_0806c8ca();
  if (in_EAX[4] != (string)0x0) {
    return in_EAX;
  }
  FUN_080641c2();
  if (DAT_0807c7b0 == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806f46b;
  }
  else if (-1 < DAT_0807c7b0 + -1) {
LAB_0806f46b:
    if ((in_EDX[0xc] == 0) && (in_EDX[0xd] == 0)) {
      poVar5 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      poVar5 = (ostream *)operator<<(poVar5,"\"pkex/pkex.cpp\":276: ECF\x1b[22m params.sendCb: ");
      endl(poVar5);
      FUN_0806b4e6("!!");
    }
  }
  if (DAT_0807c7b4 == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806f4c5;
  }
  else if (-1 < DAT_0807c7b4 + -1) {
LAB_0806f4c5:
    if ((in_EDX[0xe] == 0) && (in_EDX[0xf] == 0)) {
      piVar7 = in_EDX;
      piVar9 = in_EDX;
      poVar5 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      puVar12 = &DAT_08077b58;
      poVar5 = (ostream *)operator<<(poVar5,"\"pkex/pkex.cpp\":277: ECF\x1b[22m params.resultCb: ");
      endl(poVar5);
      FUN_0806b4e6("!!",puVar12,piVar7,piVar9);
    }
  }
  if (DAT_0807c7b8 == 1) {
    if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0806f51f;
  }
  else if (-1 < DAT_0807c7b8 + -1) {
LAB_0806f51f:
    if ((in_EDX[0x10] == 0) && (in_EDX[0x11] == 0)) {
      poVar5 = (ostream *)operator<<((ostream *)cerr,"\x1b[1m");
      poVar5 = (ostream *)operator<<(poVar5,"\"pkex/pkex.cpp\":278: ECF\x1b[22m params.doneCb: ");
      endl(poVar5);
      FUN_0806b4e6("!!");
    }
  }
  FUN_0806a60e();
  if (*in_EDX == 1) {
    puVar6 = malloc(0xe4);
    iVar1 = in_EDX[0xe];
    local_78 = (int *)in_EDX[0xf];
    if ((local_78 != (int *)0x0) && (iVar1 == 0)) {
      local_78 = (int *)(**(code **)(*local_78 + 4))(local_78);
    }
    iVar2 = in_EDX[0xc];
    piVar7 = (int *)in_EDX[0xd];
    if ((piVar7 != (int *)0x0) && (iVar2 == 0)) {
      piVar7 = (int *)(**(code **)(*piVar7 + 4))(piVar7);
    }
    iVar8 = in_EDX[1];
    iVar3 = in_EDX[2];
    FUN_080533a6();
    FUN_080533a6();
    puVar6[1] = iVar8;
    puVar6[2] = iVar3;
    *puVar6 = &DAT_0807a2d8;
    FUN_080533a6();
    piVar9 = in_EDX + 3;
    piVar10 = puVar6 + 3;
    for (iVar8 = 9; iVar8 != 0; iVar8 = iVar8 + -1) {
      *piVar10 = *piVar9;
      piVar9 = piVar9 + (uint)bVar11 * -2 + 1;
      piVar10 = piVar10 + (uint)bVar11 * -2 + 1;
    }
    puVar6[0xd] = piVar7;
    puVar6[0xc] = iVar2;
    puVar6[0x10] = &LAB_0806b52c;
    puVar6[0xe] = iVar1;
    puVar6[0xf] = local_78;
    puVar6[0x11] = in_EDX;
    puVar6[0x12] = 0;
    puVar6[0x13] = 0;
    puVar6[0x14] = 0;
    puVar6[0x15] = 0;
    puVar6[0x16] = 0;
    local_24 = 0;
    local_20 = 0;
    FUN_0806f38c();
    puVar6[0x28] = 0;
    puVar6[0x29] = 0;
    puVar6[0x2a] = 0;
    puVar6[0x2b] = 0;
    puVar6[0x2c] = 0;
    puVar6[0x2d] = 0;
    puVar6[0x2e] = 0;
    puVar6[0x2f] = 0;
    puVar6[0x30] = 0;
    puVar6[0x31] = 0;
    FUN_0805301e();
    FUN_0806ca0e();
    FUN_080641da();
    FUN_080524f8();
    *puVar6 = &PTR_FUN_0807a320;
    puVar6[0x32] = 0;
    puVar6[0x33] = 0;
    puVar6[0x34] = 0;
    puVar6[0x35] = 0;
    puVar6[0x36] = 0;
    puVar6[0x37] = 0;
    puVar6[0x38] = 0;
  }
  else {
    if (*in_EDX != 2) {
      pcVar13 = "invalid role";
      goto LAB_0806f963;
    }
    puVar6 = malloc(0xd8);
    iVar1 = in_EDX[0xe];
    local_78 = (int *)in_EDX[0xf];
    if ((iVar1 == 0) && (local_78 != (int *)0x0)) {
      local_78 = (int *)(**(code **)(*local_78 + 4))(local_78);
    }
    iVar2 = in_EDX[0xc];
    piVar7 = (int *)in_EDX[0xd];
    if ((piVar7 != (int *)0x0) && (iVar2 == 0)) {
      piVar7 = (int *)(**(code **)(*piVar7 + 4))(piVar7);
    }
    iVar8 = in_EDX[1];
    iVar3 = in_EDX[2];
    FUN_080533a6();
    FUN_080533a6();
    puVar6[1] = iVar8;
    puVar6[2] = iVar3;
    *puVar6 = &DAT_0807a2d8;
    FUN_080533a6();
    piVar9 = in_EDX + 3;
    piVar10 = puVar6 + 3;
    for (iVar8 = 9; iVar8 != 0; iVar8 = iVar8 + -1) {
      *piVar10 = *piVar9;
      piVar9 = piVar9 + (uint)bVar11 * -2 + 1;
      piVar10 = piVar10 + (uint)bVar11 * -2 + 1;
    }
    puVar6[0xd] = piVar7;
    puVar6[0xc] = iVar2;
    puVar6[0x10] = &LAB_0806b52c;
    puVar6[0xe] = iVar1;
    puVar6[0xf] = local_78;
    puVar6[0x11] = in_EDX;
    puVar6[0x12] = 0;
    puVar6[0x13] = 0;
    puVar6[0x14] = 0;
    puVar6[0x15] = 0;
    puVar6[0x16] = 0;
    local_24 = 0;
    local_20 = 0;
    FUN_0806f38c();
    puVar6[0x28] = 0;
    puVar6[0x29] = 0;
    puVar6[0x2a] = 0;
    puVar6[0x2b] = 0;
    puVar6[0x2c] = 0;
    puVar6[0x2d] = 0;
    puVar6[0x2e] = 0;
    puVar6[0x2f] = 0;
    puVar6[0x30] = 0;
    puVar6[0x31] = 0;
    FUN_0805301e();
    FUN_0806ca0e();
    FUN_080641da();
    FUN_080524f8();
    *puVar6 = &PTR_FUN_0807a2fc;
    puVar6[0x32] = 0;
    puVar6[0x33] = 0;
    puVar6[0x34] = 0;
    puVar6[0x35] = 0;
  }
  FUN_0805301e();
  FUN_0806ca0e();
  FUN_080641da();
  FUN_080524f8();
  FUN_0806a60e();
  FUN_080524f8();
  FUN_0806d06e();
  cVar4 = (**(code **)(*(int *)in_EDX[0x14] + 8))((int *)in_EDX[0x14]);
  if (cVar4 != '\0') {
    *in_EAX = (string)0x0;
    in_EAX[4] = (string)0x0;
    return in_EAX;
  }
  FUN_0806a60e();
  pcVar13 = "failed to init PKEX";
LAB_0806f963:
  string::string((string *)&local_24,pcVar13);
  string::string(in_EAX,(string *)&local_24);
  in_EAX[4] = (string)0x1;
  string::freeptr();
  return in_EAX;
}


