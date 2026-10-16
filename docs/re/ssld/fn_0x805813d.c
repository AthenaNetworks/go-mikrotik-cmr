
void FUN_0805813d(undefined1 *param_1,int param_2,undefined1 *param_3,int param_4,undefined4 param_5
                 ,undefined4 param_6)

{
  int iVar1;
  void *in_EAX;
  int iVar2;
  uint in_EDX;
  undefined4 *puVar3;
  undefined1 *puVar4;
  byte bVar5;
  undefined4 uStackY_180;
  undefined1 auStack_16c [8];
  int local_164;
  undefined1 *local_160;
  undefined **local_158 [36];
  undefined4 local_c8;
  undefined4 local_c4;
  undefined4 local_c0;
  
  bVar5 = 0;
  local_c8 = 0x8066e90;
  local_164 = param_2;
  local_c4 = 0;
  local_c0 = 0;
  uStackY_180 = 0x8058182;
  HMAC::HMAC((HMAC *)local_158,(Hash *)&local_c8,in_EAX,in_EDX);
  local_160 = auStack_16c;
  local_158[0] = &PTR_LAB_0806544c;
  iVar1 = -(param_4 + local_164 + 0xfU & 0xfffffff0);
  puVar4 = auStack_16c + iVar1;
  for (iVar2 = local_164; iVar2 != 0; iVar2 = iVar2 + -1) {
    *puVar4 = *param_1;
    param_1 = param_1 + (uint)bVar5 * -2 + 1;
    puVar4 = puVar4 + (uint)bVar5 * -2 + 1;
  }
  for (; param_4 != 0; param_4 = param_4 + -1) {
    *puVar4 = *param_3;
    param_3 = param_3 + (uint)bVar5 * -2 + 1;
    puVar4 = puVar4 + (uint)bVar5 * -2 + 1;
  }
  *(undefined1 **)((int)local_158 + iVar1 + -0x18) = param_3;
  *(undefined1 **)((int)local_158 + iVar1 + -0x1c) = param_3;
  *(undefined4 *)((int)local_158 + iVar1 + -0x20) = param_6;
  *(undefined4 *)(&stack0xfffffe84 + iVar1) = param_5;
  *(undefined4 *)((int)&uStackY_180 + iVar1) = 0x80581c9;
  FUN_08057dac();
  puVar4 = local_160;
  local_158[0] = &PTR_LAB_0806544c;
  puVar3 = (undefined4 *)(local_160 + -0x10);
  *(HMAC **)(local_160 + -0x10) = (HMAC *)local_158;
  *(undefined4 *)(puVar4 + -0x14) = 0x80581e2;
  HMAC::~HMAC((HMAC *)*puVar3);
  return;
}


