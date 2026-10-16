
void FUN_0805fa50(int *param_1,undefined4 param_2,undefined4 param_3,undefined4 param_4)

{
  char cVar1;
  undefined4 uVar2;
  uint uVar3;
  undefined2 *puVar4;
  short in_CX;
  int *piVar5;
  Rng *in_EDX;
  int iVar6;
  uint uVar7;
  undefined1 *puVar8;
  undefined1 *puVar9;
  byte bVar10;
  BigNum *pBVar11;
  BigNum *pBVar12;
  undefined1 auStack_dc [12];
  undefined4 local_d0;
  short local_ca;
  int local_c8;
  int local_c4;
  int *local_c0;
  undefined1 *local_b8;
  int local_b4;
  BigNum local_b0 [8];
  int local_a8;
  int local_a4;
  BigNum local_a0 [8];
  BigNum local_98 [8];
  BigNum local_90 [8];
  BigNum local_88 [8];
  undefined1 local_80 [8];
  BigNum local_78 [8];
  undefined1 local_70 [8];
  BigNum *local_68;
  undefined4 local_64;
  undefined4 *local_58;
  undefined4 local_54;
  undefined4 local_40;
  undefined4 local_3c;
  undefined4 local_38;
  undefined4 local_34;
  
  bVar10 = 0;
  local_b8 = (undefined1 *)0x0;
  local_b4 = 0;
  piVar5 = (int *)((int)param_1 + *(int *)(*param_1 + -0x28));
  local_ca = in_CX;
  local_c8 = (**(code **)(*piVar5 + 0x18))(piVar5);
  if (local_c8 == 0) {
    calcDigest(&local_40);
    RsaPrivateKey::signHashed(&local_58,param_1 + -5,param_2,local_40,local_3c);
    puVar8 = &stack0xffffff1c;
    FUN_0805091c();
  }
  else {
    if (local_c8 == 2) {
      local_d0 = param_4;
      FUN_08050d68();
      local_68 = local_78;
      local_78[0] = (BigNum)0x1;
      local_64 = 1;
      FUN_0804f6aa();
      local_c0 = param_1 + -4;
      FUN_08050e36();
      BigNum::operator-=((BigNum *)&local_40,(BigNum *)&local_58);
      uVar2 = randomBigNum(local_b0,(BigNum *)&local_40,in_EDX);
      FUN_0804edbe((BigNum *)&local_40);
      FUN_0804edbe((BigNum *)&local_58);
      puVar8 = auStack_dc;
      if ((char)uVar2 == '\0') {
        local_40 = 0;
        local_3c = 0;
      }
      else {
        calcDigest(&local_a8);
        uVar7 = (local_a4 - local_a8) * 8;
        BigNum::BigNum(local_a0,local_a8,local_a4,uVar2,uVar2);
        uVar3 = BigNum::bits();
        if (uVar3 < uVar7) {
          BigNum::operator>>=(local_a0,uVar7 - uVar3);
        }
        FUN_0805b76a();
        FUN_08059380(param_1 + -8);
        WCurve::toEC((Point *)&local_68);
        puVar8 = &stack0xffffff1c;
        FUN_08059352();
        FUN_0805ef02();
        cVar1 = FUN_0805e7ca();
        if (cVar1 == '\0') {
          invByMod(local_90,local_b0);
          uVar2 = FUN_08050d68();
          pBVar12 = (BigNum *)(param_1 + 1);
          mul(local_78,local_98,pBVar12);
          FUN_08050e36();
          pBVar11 = local_78;
          BigNum::operator+=((BigNum *)&local_40,local_78);
          FUN_0805ef02();
          FUN_0804edbe((BigNum *)&local_40,pBVar11,pBVar12,uVar2);
          FUN_0804edbe(local_78);
          FUN_08050d68();
          mul((BigNum *)&local_40,local_90,local_88);
          FUN_0805ef02();
          FUN_0804edbe((BigNum *)&local_40);
          FUN_0804fbdf();
          FUN_0804fbdf();
          cVar1 = FUN_0805e7ca();
          if (cVar1 == '\0') {
            FUN_08050e36();
            FUN_08050e36();
          }
          else {
            FUN_08050d68();
            FUN_08050d68();
          }
          FUN_0804edbe(local_80);
          FUN_0804edbe(local_88);
          FUN_0804edbe(local_90);
          puVar8 = &stack0xffffff18;
        }
        else {
          FUN_08050d68();
          FUN_08050d68();
        }
        *(BigNum **)(puVar8 + -0x10) = local_98;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fd74;
        FUN_0804edbe();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fd7c;
        FUN_0804f6cc();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fd84;
        FUN_0805931c();
        *(BigNum **)(puVar8 + -0x10) = local_a0;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fd92;
        FUN_0804edbe();
        local_40 = 0;
        local_3c = 0;
        local_38 = 0;
        local_34 = 0;
        *(undefined4 *)(puVar8 + -8) = 0;
        *(undefined4 *)(puVar8 + -0xc) = 0x30;
        *(undefined4 **)(puVar8 + -0x10) = &local_40;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdad;
        DerEncoder::addTag(*(uint *)(puVar8 + -0x10),*(uint *)(puVar8 + -0xc));
        local_54 = 0;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdc0;
        local_58 = &local_40;
        FUN_0804ffa8();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdc8;
        FUN_0804f6ee();
        *(BigNum **)(puVar8 + -8) = local_78;
        *(undefined4 *)(puVar8 + -0xc) = 2;
        *(undefined4 **)(puVar8 + -0x10) = &local_40;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdd4;
        uVar2 = DerEncoder::addInt(*(uint *)(puVar8 + -0x10),*(BigNum **)(puVar8 + -0xc));
        *(undefined1 **)(puVar8 + -8) = local_70;
        *(undefined4 *)(puVar8 + -0xc) = 2;
        *(undefined4 *)(puVar8 + -0x10) = uVar2;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fde0;
        local_c0 = (int *)DerEncoder::addInt(*(uint *)(puVar8 + -0x10),*(BigNum **)(puVar8 + -0xc));
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdf0;
        FUN_0805001a();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fdfe;
        FUN_080504a6();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe06;
        FUN_0804ff86();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe0d;
        FUN_0805091c();
        *(undefined1 **)(puVar8 + -0x10) = local_70;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe15;
        FUN_0804edbe();
        *(BigNum **)(puVar8 + -0x10) = local_78;
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe1d;
        FUN_0804edbe();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe28;
        FUN_0805091c();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe33;
        FUN_0804fbdf();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe3d;
        FUN_0804ff50();
        *(undefined4 *)(puVar8 + -0x14) = 0x805fe45;
        FUN_0805091c();
      }
      *(BigNum **)(puVar8 + -0x10) = local_b0;
      *(undefined4 *)(puVar8 + -0x14) = 0x805fe57;
      FUN_0804edbe();
      *(undefined4 *)(puVar8 + -0x14) = 0x805fe65;
      FUN_08050a6a();
      *(undefined4 *)(puVar8 + -0x14) = 0x805fe6d;
      FUN_0805091c();
      goto LAB_0805fec3;
    }
    if (local_c8 != 3) {
                    /* WARNING: Subroutine does not return */
      __assert_fail("false","tls-msgs.cpp",0x80,"signAndAppendDigitallySigned");
    }
    Ed25519PrivateKey::sign(&local_40);
    puVar8 = &stack0xffffff20;
  }
  *(undefined4 *)(puVar8 + -4) = 0x805fe9e;
  FUN_08050a6a();
  *(undefined4 *)(puVar8 + -4) = 0x805fea5;
  FUN_0805091c();
LAB_0805fec3:
  if ((local_ca == -0x103) || (local_ca == 0x303)) {
    *(undefined4 *)(puVar8 + -4) = 0x805ff09;
    FUN_08058f1a();
    puVar4 = (undefined2 *)((-4 - (local_b4 - (int)local_b8)) + *(int *)(local_c4 + 4));
    if (local_c8 == 3) {
      *puVar4 = 0x708;
    }
    else {
      switch(param_2) {
      case 3:
        *(undefined1 *)puVar4 = 1;
        break;
      case 4:
        *(undefined1 *)puVar4 = 2;
        break;
      case 5:
        *(undefined1 *)puVar4 = 4;
        break;
      case 6:
        *(undefined1 *)puVar4 = 5;
        break;
      case 7:
        *(undefined1 *)puVar4 = 6;
      }
      if (local_c8 == 0) {
        *(undefined1 *)((int)puVar4 + 1) = 1;
      }
      else if (local_c8 == 2) {
        *(undefined1 *)((int)puVar4 + 1) = 3;
      }
    }
  }
  else {
    *(undefined4 *)(puVar8 + -4) = 0x805ff8b;
    FUN_08058f1a();
  }
  iVar6 = local_b4 - (int)local_b8;
  puVar4 = (undefined2 *)((-2 - iVar6) + *(int *)(local_c4 + 4));
  *puVar4 = CONCAT11((char)iVar6,(char)((uint)iVar6 >> 8));
  puVar9 = local_b8;
  puVar4 = puVar4 + 1;
  for (iVar6 = local_b4 - (int)local_b8; iVar6 != 0; iVar6 = iVar6 + -1) {
    *(undefined1 *)puVar4 = *puVar9;
    puVar9 = puVar9 + (uint)bVar10 * -2 + 1;
    puVar4 = (undefined2 *)((int)puVar4 + (uint)bVar10 * -2 + 1);
  }
  *(undefined4 *)(puVar8 + -4) = 0x805ffca;
  FUN_0805091c();
  return;
}


