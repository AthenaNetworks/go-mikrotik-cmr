
void FUN_08057dac(undefined1 *param_1,uint param_2)

{
  code *pcVar1;
  int *in_EAX;
  uint uVar2;
  int iVar3;
  int *piVar4;
  uint uVar5;
  undefined4 in_ECX;
  undefined4 extraout_ECX;
  uint uVar6;
  undefined4 in_EDX;
  undefined1 *unaff_ESI;
  byte bVar7;
  uint auStack_50 [5];
  undefined1 auStack_3c [12];
  undefined4 local_30;
  undefined4 local_2c;
  undefined1 *local_28;
  undefined1 *local_24;
  uint local_20;
  
  bVar7 = 0;
  auStack_50[0] = 0x8057dc3;
  local_30 = in_ECX;
  local_2c = in_EDX;
  local_20 = (**(code **)(*in_EAX + 0x14))();
  uVar2 = local_20 + 0xf & 0xfffffff0;
  local_24 = auStack_3c + uVar2 * -2;
  local_28 = auStack_3c + -uVar2;
  *(undefined4 *)((int)auStack_50 + uVar2 * -2 + 0x10) = extraout_ECX;
  iVar3 = *in_EAX;
  *(undefined4 *)((int)auStack_50 + uVar2 * -2 + 0xc) = local_30;
  *(undefined4 *)((int)auStack_50 + uVar2 * -2 + 8) = local_2c;
  while( true ) {
    *(int **)((int)auStack_50 + uVar2 * -2 + 4) = in_EAX;
    pcVar1 = *(code **)(iVar3 + 0xc);
    *(undefined4 *)((int)auStack_50 + uVar2 * -2) = 0x8057de9;
    piVar4 = (int *)(*pcVar1)();
    iVar3 = *piVar4;
    *(undefined1 **)((int)auStack_50 + uVar2 * -2 + 8) = local_24;
    *(int **)((int)auStack_50 + uVar2 * -2 + 4) = piVar4;
    pcVar1 = *(code **)(iVar3 + 0x18);
    *(undefined4 *)((int)auStack_50 + uVar2 * -2) = 0x8057df4;
    (*pcVar1)();
    *(undefined1 **)((int)auStack_50 + uVar2 * -2 + 0x10) = unaff_ESI;
    iVar3 = *in_EAX;
    *(uint *)((int)auStack_50 + uVar2 * -2 + 0xc) = local_20;
    *(undefined1 **)((int)auStack_50 + uVar2 * -2 + 8) = local_24;
    *(int **)((int)auStack_50 + uVar2 * -2 + 4) = in_EAX;
    pcVar1 = *(code **)(iVar3 + 0xc);
    *(undefined4 *)((int)auStack_50 + uVar2 * -2) = 0x8057e04;
    piVar4 = (int *)(*pcVar1)();
    iVar3 = *piVar4;
    *(undefined4 *)((int)auStack_50 + uVar2 * -2 + 0xc) = local_30;
    *(undefined4 *)((int)auStack_50 + uVar2 * -2 + 8) = local_2c;
    *(int **)((int)auStack_50 + uVar2 * -2 + 4) = piVar4;
    pcVar1 = *(code **)(iVar3 + 0xc);
    *(undefined4 *)((int)auStack_50 + uVar2 * -2) = 0x8057e13;
    piVar4 = (int *)(*pcVar1)();
    iVar3 = *piVar4;
    *(undefined1 **)((int)auStack_50 + uVar2 * -2 + 8) = local_28;
    *(int **)((int)auStack_50 + uVar2 * -2 + 4) = piVar4;
    pcVar1 = *(code **)(iVar3 + 0x18);
    *(undefined4 *)((int)auStack_50 + uVar2 * -2) = 0x8057e1e;
    (*pcVar1)();
    uVar5 = local_20;
    if (param_2 <= local_20) {
      uVar5 = param_2;
    }
    param_2 = param_2 - uVar5;
    unaff_ESI = local_28;
    for (uVar6 = uVar5; uVar6 != 0; uVar6 = uVar6 - 1) {
      *param_1 = *unaff_ESI;
      unaff_ESI = unaff_ESI + (uint)bVar7 * -2 + 1;
      param_1 = param_1 + (uint)bVar7 * -2 + 1;
    }
    if (param_2 == 0) break;
    *(uint *)((int)auStack_50 + uVar2 * -2 + 0x10) = uVar5;
    iVar3 = *in_EAX;
    *(uint *)((int)auStack_50 + uVar2 * -2 + 0xc) = local_20;
    *(undefined1 **)((int)auStack_50 + uVar2 * -2 + 8) = local_24;
  }
  return;
}


