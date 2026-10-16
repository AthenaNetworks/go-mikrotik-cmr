
undefined4 FUN_0806f1e4(void)

{
  undefined4 in_EAX;
  undefined4 *puVar1;
  undefined4 extraout_ECX;
  int iVar2;
  undefined4 *puVar3;
  byte bVar4;
  undefined1 *puVar5;
  undefined4 uVar6;
  undefined4 uVar7;
  Hash local_128 [8];
  undefined1 local_120 [8];
  vector local_118 [8];
  undefined4 local_110;
  undefined4 local_10c;
  undefined4 local_108 [16];
  undefined4 local_c8;
  undefined4 local_c4;
  undefined4 local_c0;
  
  bVar4 = 0;
  FUN_08051dd0();
  FUN_08051dd0();
  local_c8 = 0x807bcb0;
  local_c4 = 0;
  local_c0 = 0;
  nv::rosMode();
  FUN_08051c94();
  FUN_08051c94();
  puVar5 = local_120;
  local_110 = 0;
  local_10c = 0;
  uVar6 = 0x20;
  uVar7 = extraout_ECX;
  HKPDF(local_128,(vector *)&local_c8,(vector *)&local_110,local_118,(uint)puVar5);
  FUN_08053188(uVar7);
  FUN_080521aa(local_128,puVar5,uVar6);
  FUN_080521aa((vector *)&local_110);
  FUN_080521aa(local_118);
  FUN_080521aa(local_120);
  FUN_08051c94();
  eddsa::get_public_key(local_118,local_110,local_10c);
  FUN_08053188();
  FUN_080521aa(local_118);
  FUN_080521aa((vector *)&local_110);
  puVar1 = malloc(0x40);
  puVar3 = local_108;
  for (iVar2 = 0x10; iVar2 != 0; iVar2 = iVar2 + -1) {
    *puVar1 = *puVar3;
    puVar3 = puVar3 + (uint)bVar4 * -2 + 1;
    puVar1 = puVar1 + (uint)bVar4 * -2 + 1;
  }
  FUN_080536b2();
  return in_EAX;
}


