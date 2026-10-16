
/* WARNING: Type propagation algorithm not settling */

BigNum * FUN_0806004a(int param_1,BigNum *param_2)

{
  BigNum BVar1;
  undefined2 uVar2;
  undefined4 uVar3;
  code *pcVar4;
  void *pvVar5;
  byte bVar6;
  char cVar7;
  char cVar8;
  undefined1 uVar9;
  short sVar10;
  short sVar11;
  short *psVar12;
  ostream *poVar13;
  int *piVar14;
  BigNum *pBVar15;
  undefined4 *puVar16;
  undefined4 *puVar17;
  ostream *poVar18;
  undefined4 uVar19;
  uint uVar20;
  BigNum *extraout_ECX;
  undefined4 extraout_ECX_00;
  int iVar21;
  undefined4 extraout_ECX_01;
  undefined4 extraout_ECX_02;
  BigNum *extraout_ECX_03;
  BigNum *extraout_ECX_04;
  BigNum *extraout_ECX_05;
  BigNum *extraout_ECX_06;
  BigNum *extraout_ECX_07;
  BigNum *extraout_ECX_08;
  BigNum *extraout_ECX_09;
  BigNum *extraout_ECX_10;
  undefined4 extraout_ECX_11;
  BigNum *extraout_ECX_12;
  BigNum *extraout_ECX_13;
  byte bVar22;
  BigNum *extraout_EDX;
  BigNum *extraout_EDX_00;
  ushort uVar23;
  uint uVar24;
  BigNum *pBVar25;
  int *piVar26;
  BigNum **ppBVar27;
  undefined1 *puVar28;
  undefined1 *puVar29;
  undefined1 *puVar31;
  BigNum **ppBVar32;
  undefined1 *puVar34;
  ushort uVar35;
  BigNum *pBVar36;
  undefined1 auVar37 [4];
  BigNum *pBVar38;
  int iVar39;
  bool bVar40;
  byte bVar41;
  BigNum *pBStack_130;
  BigNum *local_12c;
  int local_128;
  undefined2 local_122;
  BigNum *local_120;
  BigNum *local_11c;
  undefined4 local_118;
  BigNum *local_114;
  BigNum *local_110;
  int *local_10c;
  undefined4 local_108;
  BigNum *local_104;
  undefined4 local_100;
  undefined4 local_fc [10];
  undefined4 local_d4;
  undefined4 local_d0;
  BigNum *local_c0;
  BigNum *local_bc;
  BigNum *local_b8;
  int local_b4;
  int local_b0;
  int local_ac;
  int local_a8;
  BigNum *local_a4;
  BigNum *local_a0;
  BigNum *local_9c;
  BigNum *local_98;
  BigNum *local_94;
  BigNum *local_90;
  BigNum *local_8c;
  BigNum *local_88;
  undefined4 local_84;
  BigNum *local_80;
  int local_7c;
  undefined4 local_78 [2];
  undefined1 local_70 [4];
  BigNum *local_6c;
  BigNum *local_68;
  BigNum *local_64;
  undefined4 local_60;
  undefined4 local_5c;
  undefined4 local_58;
  undefined2 local_54;
  undefined2 uStack_52;
  undefined2 local_50;
  undefined2 uStack_4e;
  uchar local_4c [8];
  undefined4 local_44;
  undefined4 *local_40;
  undefined4 *local_3c;
  undefined4 local_38;
  undefined4 local_34;
  undefined4 local_30;
  undefined4 local_2c;
  undefined4 local_28;
  undefined4 local_24;
  undefined4 local_20;
  undefined1 *puVar30;
  undefined1 *puVar33;
  
  bVar41 = 0;
  ppBVar32 = &local_12c;
  ppBVar27 = &local_12c;
  pBStack_130 = (BigNum *)0x8060061;
  bVar6 = FUN_080587fa();
  BVar1 = *param_2;
  pBStack_130 = (BigNum *)0x806006d;
  local_10c = (int *)FUN_080587e6();
  local_104 = param_2 + 4;
  pBVar36 = (BigNum *)(local_10c + 1);
  if (bVar6 != 0) {
    pBVar36 = (BigNum *)(local_10c + 3);
    local_104 = param_2 + 0xc;
  }
  switch(*(undefined4 *)(param_1 + 0x158)) {
  case 0:
    bVar22 = (char)BVar1 - 2;
joined_r0x0806012a:
    if (1 < bVar22) goto LAB_08060133;
    goto switchD_080600bd_default;
  case 1:
    if (*(int *)(param_1 + 0x8c) == 1) {
      bVar40 = BVar1 == (BigNum)0xc;
    }
    else {
      bVar40 = BVar1 == (BigNum)0xb;
    }
    break;
  case 2:
    bVar6 = 0;
    if (*(char *)(param_1 + 400) != '\0') {
      local_110 = (BigNum *)CONCAT31(local_110._1_3_,*(byte *)(param_1 + 0x191));
      bVar6 = (*(byte *)(param_1 + 0x191) ^ 1) & BVar1 == (BigNum)0x16;
    }
    bVar6 = bVar6 ^ 1;
    if (*(int *)(param_1 + 0x88) == 0) {
      bVar22 = (char)BVar1 - 0xd;
      goto joined_r0x0806012a;
    }
    if (BVar1 == (BigNum)0xc) goto switchD_080600bd_default;
LAB_08060133:
    bVar40 = bVar6 == 0;
    break;
  case 3:
    if (*(int *)(param_1 + 0x8c) == 1) goto switchD_080600bd_caseD_4;
    if ((byte)((char)BVar1 - 0xdU) < 2) goto switchD_080600bd_default;
    goto switchD_080600bd_caseD_5;
  case 4:
switchD_080600bd_caseD_4:
    bVar40 = BVar1 == (BigNum)0xe;
    break;
  case 5:
    goto switchD_080600bd_caseD_5;
  case 6:
    if (BVar1 != (BigNum)0x14) goto switchD_080600bd_caseD_5;
    goto LAB_080601a2;
  default:
    goto switchD_080600bd_default;
  }
  if (bVar40) {
switchD_080600bd_default:
    if ((BVar1 != (BigNum)0x14) && (BVar1 != (BigNum)0x3)) {
      pBStack_130 = (BigNum *)0x80601a2;
      FUN_080594ae();
      pBVar36 = extraout_ECX;
    }
LAB_080601a2:
    switch(BVar1) {
    case (BigNum)0x2:
      pBStack_130 = (BigNum *)0x806034d;
      cVar7 = FUN_080587fa();
      if (cVar7 != '\0') {
        pBStack_130 = (BigNum *)0x8060360;
        FUN_08050912();
        *(undefined4 *)(param_1 + 0x140) = 0;
        *(undefined4 *)(param_1 + 0x144) = 100;
      }
      auVar37 = local_70;
      local_4c[0] = '\0';
      local_4c[1] = '\0';
      local_4c[2] = '\0';
      local_4c[3] = '\0';
      local_4c[4] = '\0';
      local_4c[5] = '\0';
      local_4c[6] = '\0';
      local_4c[7] = '\0';
      local_40 = (undefined4 *)0x0;
      local_3c = (undefined4 *)0x0;
      if (local_10c < (int *)0x26) {
LAB_08060436:
        pBStack_130 = param_2;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," parsing ServerHello");
        endl(poVar13);
        string::freeptr();
        local_80 = (BigNum *)0x11;
        string::string((string *)&local_7c);
        FUN_0805a4bc();
      }
      else {
        local_70[1] = (char)*(undefined2 *)local_104;
        local_70[0] = (char)((ushort)*(undefined2 *)local_104 >> 8);
        local_70._2_2_ = auVar37._2_2_;
        uStack_52 = (undefined2)*(undefined4 *)(local_104 + 0x1e);
        local_50 = (undefined2)((uint)*(undefined4 *)(local_104 + 0x1e) >> 0x10);
        pBVar36 = local_104 + 2;
        puVar16 = (undefined4 *)(local_70 + 2);
        for (iVar21 = 7; iVar21 != 0; iVar21 = iVar21 + -1) {
          *puVar16 = *(undefined4 *)pBVar36;
          pBVar36 = pBVar36 + (uint)bVar41 * -8 + 4;
          puVar16 = puVar16 + (uint)bVar41 * -2 + 1;
        }
        param_2 = (BigNum *)(uint)(byte)local_104[0x22];
        if (((BigNum *)0x20 < param_2) || ((BigNum *)((int)local_10c - 0x23U) < param_2 + 3))
        goto LAB_08060436;
        pBVar36 = (BigNum *)((int)local_10c + (int)local_104);
        param_2 = param_2 + (int)(local_104 + 0x23);
        pBStack_130 = param_2;
        vector<unsigned_char>::insert(local_4c,(uchar *)0x0,(uchar *)(local_104 + 0x23));
        local_44._0_3_ =
             CONCAT12(param_2[2],
                      CONCAT11((char)*(undefined2 *)param_2,
                               (char)((ushort)*(undefined2 *)param_2 >> 8)));
        if (pBVar36 != param_2 + 3) {
          if (((int)pBVar36 - (int)(param_2 + 3) != 1) &&
             ((uint)CONCAT11((char)*(undefined2 *)(param_2 + 3),
                             (char)((ushort)*(undefined2 *)(param_2 + 3) >> 8)) ==
              (int)pBVar36 - (int)(param_2 + 5))) {
            pBStack_130 = (BigNum *)0x8060432;
            cVar7 = FUN_0805f8b4();
            if (cVar7 != '\0') goto LAB_0806049b;
          }
          goto LAB_08060436;
        }
LAB_0806049b:
        pBStack_130 = pBVar36;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        operator<<(poVar13," rx:");
        poVar13 = (ostream *)FUN_0805977e();
        endl(poVar13);
        string::freeptr();
        auVar37 = local_70;
        cVar7 = FUN_080587fa();
        puVar16 = local_3c;
        iVar21 = *(int *)(param_1 + 0x30);
        uVar35 = (ushort)*(undefined4 *)(iVar21 + 8);
        uVar23 = auVar37._0_2_;
        if (cVar7 == '\0') {
          if (uVar35 <= uVar23) {
            bVar40 = *(ushort *)(iVar21 + 10) < uVar23;
            goto LAB_08060517;
          }
        }
        else if ((uVar23 <= uVar35) && (uVar23 != 0xfefe)) {
          bVar40 = uVar23 < *(ushort *)(iVar21 + 10);
LAB_08060517:
          puVar17 = local_40;
          if (!bVar40) {
LAB_08060549:
            if (puVar16 != puVar17) goto code_r0x0806054d;
            pBStack_130 = (BigNum *)0x80605b0;
            iVar21 = FUN_08059468();
            if ((iVar21 != 0) && (*(int *)(iVar21 + 8) - *(int *)(iVar21 + 4) == 4)) {
              pBStack_130 = (BigNum *)0x80605cc;
              FUN_0804f222();
            }
            if (*(int *)(*(int *)(param_1 + 0x30) + 0xc) == 1) {
              pBStack_130 = (BigNum *)0x80605e5;
              iVar21 = FUN_08059468();
              if ((iVar21 != 0) && (*(char *)(iVar21 + 4) == '\x01')) {
                *(undefined1 *)(param_1 + 400) = 1;
              }
            }
            psVar12 = *(short **)(*(int *)(param_1 + 0x15c) + 0x34);
            while (*(short **)(*(int *)(param_1 + 0x15c) + 0x38) != psVar12) {
              sVar11 = *psVar12;
              psVar12 = psVar12 + 1;
              if ((short)local_44 == sVar11) {
                *(short *)(param_1 + 0x80) = local_70._0_2_;
                local_80 = (BigNum *)0x0;
                local_7c = 0;
                pBStack_130 = (BigNum *)&uStack_4e;
                vector<unsigned_char>::insert((uchar *)&local_80,(uchar *)0x0,local_70 + 2);
                FUN_08050a6a();
                FUN_0805091c();
                FUN_0805942e();
                puVar16 = local_fc;
                puVar17 = (undefined4 *)(param_1 + 0x84);
                for (iVar21 = 7; iVar21 != 0; iVar21 = iVar21 + -1) {
                  *puVar17 = *puVar16;
                  puVar16 = puVar16 + (uint)bVar41 * -2 + 1;
                  puVar17 = puVar17 + (uint)bVar41 * -2 + 1;
                }
                sVar11 = (short)*(undefined4 *)(param_1 + 0x80);
                if (((sVar11 == -0x103) || (uVar19 = 1, sVar11 == 0x303)) &&
                   (uVar19 = 5, *(int *)(param_1 + 0x9c) == 3)) {
                  uVar19 = 6;
                }
                *(undefined4 *)(param_1 + 0x50) = uVar19;
                pBStack_130 = (BigNum *)&local_80;
                FUN_0805986a();
                pvVar5 = *(void **)(param_1 + 0x15c);
                *(undefined4 *)(param_1 + 0x15c) = 0;
                if (pvVar5 != (void *)0x0) {
                  pBStack_130 = (BigNum *)0x80628ee;
                  FUN_08051b50();
                  free(pvVar5);
                }
                *(undefined4 *)(param_1 + 0x158) = 1;
                pBStack_130 = (BigNum *)0x8062890;
                FUN_08051a20();
                pBStack_130 = (BigNum *)0x8062898;
                FUN_0805091c();
                return (BigNum *)0x1;
              }
            }
            local_80 = (BigNum *)0x12;
            string::string((string *)&local_7c);
            goto LAB_080628ba;
          }
        }
        local_80 = (BigNum *)0x29;
        *(ushort *)(param_1 + 0x80) = uVar23;
        string::string((string *)&local_7c);
LAB_080628ba:
        FUN_0805a4bc();
      }
      string::freeptr();
      ppBVar32 = &local_12c;
      pBStack_130 = (BigNum *)0x80628d5;
      FUN_08051a20();
      break;
    case (BigNum)0x3:
      pBStack_130 = (BigNum *)0x80601bc;
      cVar7 = FUN_080587fa();
      if (cVar7 != '\0') {
        pBStack_130 = (BigNum *)0x80601cf;
        FUN_08050912();
        *(undefined4 *)(param_1 + 0x140) = 0;
        *(undefined4 *)(param_1 + 0x144) = 100;
      }
      auVar37 = local_70;
      local_6c = (BigNum *)0x0;
      local_68 = (BigNum *)0x0;
      if (local_10c < (int *)0x3) {
LAB_080602d0:
        pBStack_130 = param_2;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," parsing HelloVerifyRequest");
        endl(poVar13);
        string::freeptr();
        local_80 = (BigNum *)0x11;
        string::string((string *)&local_7c);
        FUN_0805a4bc();
        string::freeptr();
        FUN_0805091c();
        return (BigNum *)0x0;
      }
      local_70[1] = (char)*(undefined2 *)local_104;
      local_70[0] = (char)((ushort)*(undefined2 *)local_104 >> 8);
      local_70._2_2_ = auVar37._2_2_;
      if ((uint)(byte)local_104[2] != (int)local_10c - 3U) goto LAB_080602d0;
      pBStack_130 = local_104 + 3 + (byte)local_104[2];
      puVar34 = &stack0xfffffec4;
      vector<unsigned_char>::insert((uchar *)&local_6c,(uchar *)0x0,(uchar *)(local_104 + 3));
      local_104 = *(BigNum **)(param_1 + 0x15c);
      FUN_08051908();
      *(short *)(param_1 + 0x120) = *(short *)(param_1 + 0x120) + 1;
      FUN_080587fa();
      FUN_0805ec4c();
      FUN_08050912();
      FUN_080594e2();
      FUN_0805a416();
      FUN_08059e70();
      FUN_0805091c();
      goto LAB_0806260d;
    default:
      goto switchD_080600bd_caseD_5;
    case (BigNum)0xb:
      piVar26 = (int *)&stack0xfffffec4;
      puVar16 = malloc(8);
      *puVar16 = 0;
      puVar16[1] = 0;
      local_70 = (undefined1  [4])0x0;
      FUN_080519fa();
      FUN_080519fa();
      pBVar36 = (BigNum *)FUN_0805ea8e();
      ppBVar32 = &local_12c;
      if ((char)pBVar36 == '\0') {
        pBStack_130 = pBVar36;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," parsing Certificate");
        endl(poVar13);
        string::freeptr();
        local_70 = (undefined1  [4])0x11;
        string::string((string *)&local_6c);
        pBStack_130 = (BigNum *)0x80606f2;
        FUN_0805a4bc();
        goto LAB_0806017c;
      }
      if (**(int **)(param_1 + 0xec) == (*(int **)(param_1 + 0xec))[1]) {
        local_70 = (undefined1  [4])0x17;
        string::string((string *)&local_6c);
        goto LAB_080619de;
      }
      if (*(int *)(*(int *)(param_1 + 0x30) + 0xc) == 1) {
        local_80 = (BigNum *)0x0;
        local_7c = 0;
        FUN_080521be();
        FUN_080504c6();
        x509BuildPath();
        FUN_08050ebc();
        string::string((string *)&local_8c);
        local_110 = (BigNum *)
                    CONCAT31(local_110._1_3_,*(undefined1 *)(*(int *)(param_1 + 0x30) + 0x30));
        local_10c = *(undefined4 **)(param_1 + 0x174) + 1;
        local_108 = **(undefined4 **)(param_1 + 0x174);
        local_104 = (BigNum *)(*(undefined4 **)(param_1 + 0x170) + 1);
        local_100 = **(undefined4 **)(param_1 + 0x170);
        TimePoint::now();
        iVar21 = x509ValidatePath(&local_80,*(undefined4 *)(*(int *)(param_1 + 0x30) + 0x28),
                                  local_d4,local_d0,local_104,local_100,param_1 + 0x178,local_10c,
                                  local_108,(uint)local_110 & 0xff,(string *)&local_8c);
        if (iVar21 != 0) {
          local_104 = (BigNum *)(iVar21 - 1);
          local_70 = (undefined1  [4])0x0;
          if (local_104 < 0xd) {
            local_70 = (undefined1  [4])(int)(char)(&DAT_080650f3)[iVar21];
          }
          local_10c = (int *)iVar21;
          string::string((string *)&local_6c,(string *)&local_8c);
          FUN_0805a4bc();
          string::freeptr();
          string::freeptr();
          FUN_08050f2a();
          return (BigNum *)0x0;
        }
        string::freeptr();
        FUN_08050f2a();
        ppBVar32 = &pBStack_130;
      }
      *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x80608bb;
      FUN_0805181c();
      uVar19 = *(undefined4 *)((int)local_70 + 0x74);
      uVar3 = *(undefined4 *)((int)local_70 + 0x78);
      *(undefined1 (*) [4])((int)ppBVar32 + 0xfffffffc) = local_70;
      *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = uVar3;
      *(undefined4 *)((int)ppBVar32 + 0xfffffff4) = uVar19;
      *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_80;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x80608cd;
      parsePkcs8PublicKey();
      local_80 = (BigNum *)0x0;
      *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 0x80608e3;
      FUN_08051856();
      *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 0x80608ea;
      FUN_08051872();
      *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 0x80608f2;
      FUN_08050ebc();
      piVar14 = *(int **)(param_1 + 0x164);
      if (piVar14 != (int *)0x0) {
        if (*(int *)(param_1 + 0x8c) == 0) {
          piVar14 = (int *)((int)piVar14 + *(int *)(*piVar14 + -0x28));
          iVar21 = *piVar14;
          *(int **)((int)ppBVar32 + 0xffffffec) = piVar14;
          pcVar4 = *(code **)(iVar21 + 0x18);
          *(undefined4 *)((int)ppBVar32 + 0xffffffe8) = 0x8060919;
          iVar21 = (*pcVar4)();
          if (iVar21 != 0) goto LAB_08060976;
        }
        if (*(int *)(param_1 + 0x8c) != 2) {
LAB_08060996:
          *(undefined4 *)(param_1 + 0x158) = 2;
          return pBVar36;
        }
        piVar14 = (int *)((int)*(int **)(param_1 + 0x164) +
                         *(int *)(**(int **)(param_1 + 0x164) + -0x28));
        iVar21 = *piVar14;
        *(int **)((int)ppBVar32 + 0xffffffec) = piVar14;
        pcVar4 = *(code **)(iVar21 + 0x18);
        *(undefined4 *)((int)ppBVar32 + 0xffffffe8) = 0x8060940;
        iVar21 = (*pcVar4)();
        piVar14 = *(int **)(param_1 + 0x164);
        if (iVar21 == 2) {
          if (piVar14[5] - 3U < 3) goto LAB_08060996;
        }
        else {
          piVar14 = (int *)((int)piVar14 + *(int *)(*piVar14 + -0x28));
          iVar21 = *piVar14;
          *(int **)((int)ppBVar32 + 0xffffffec) = piVar14;
          pcVar4 = *(code **)(iVar21 + 0x18);
          *(undefined4 *)((int)ppBVar32 + 0xffffffe8) = 0x806096e;
          iVar21 = (*pcVar4)();
          if (iVar21 == 3) goto LAB_08060996;
        }
      }
LAB_08060976:
      local_70 = (undefined1  [4])0x1d;
      piVar26 = (int *)((int)ppBVar32 + 0xffffffec);
      *(BigNum ***)((int)ppBVar32 + 0xffffffec) = &local_6c;
      *(undefined4 *)((int)ppBVar32 + 0xffffffe8) = 0x8060989;
      string::string(*(string **)((int)ppBVar32 + 0xffffffec));
      goto LAB_08060171;
    case (BigNum)0xc:
      local_8c = (BigNum *)0x0;
      local_88 = (BigNum *)0x0;
      local_84 = 0;
      local_110 = (BigNum *)((int)local_10c + (int)local_104);
      ppBVar32 = &local_12c;
      if (local_104 == local_110) goto LAB_080615ed;
      pBVar36 = *(BigNum **)(param_1 + 0x164);
      local_122 = (undefined2)*(undefined4 *)(param_1 + 0x80);
      local_128 = *(int *)(param_1 + 0x8c);
      local_11c = pBVar36;
      if (*(int *)(param_1 + 0x88) == 1) {
        local_c0 = malloc(0x1c);
        pBVar36 = local_c0;
        for (iVar21 = 7; iVar21 != 0; iVar21 = iVar21 + -1) {
          *(undefined4 *)pBVar36 = 0;
          pBVar36 = pBVar36 + (uint)bVar41 * -8 + 4;
        }
        *(undefined **)local_c0 = &DAT_08064640;
        FUN_080505f8();
        *(undefined ***)local_c0 = &PTR_FUN_08064cf4;
        FUN_08050d68();
        FUN_08050d68();
        FUN_08050d68();
        local_b4 = 0;
        local_bc = local_104;
        local_b0 = 0;
        local_b8 = local_110;
        local_ac = 0;
        local_a8 = 0;
        local_a4 = (BigNum *)0x0;
        local_a0 = (BigNum *)0x0;
        cVar7 = FUN_0805e7f7();
        if (cVar7 != '\0') {
          pBStack_130 = (BigNum *)0x8060fb8;
          cVar7 = FUN_0805e7f7();
          if (cVar7 != '\0') {
            pBStack_130 = (BigNum *)0x8060fd1;
            cVar7 = FUN_0805e7f7();
            if ((((cVar7 != '\0') && (local_b4 != local_b0)) && (local_ac != local_a8)) &&
               (local_a4 != local_a0)) {
              local_114 = local_bc;
              local_110 = local_b8;
              pBStack_130 = param_2;
              BigNum::BigNum();
              uVar20 = BigNum::bits();
              if ((0x7ff < uVar20) && (uVar20 = BigNum::bits(), pBVar36 = local_c0, uVar20 < 0x2001)
                 ) {
                pBStack_130 = extraout_ECX_03;
                BigNum::BigNum();
                BigNum::BigNum();
                local_70 = (undefined1  [4])0x0;
                local_6c = (BigNum *)0x0;
                (*assertRelated)();
                local_6c = local_98;
                local_70 = (undefined1  [4])local_9c;
                local_9c = (BigNum *)0x0;
                local_98 = (BigNum *)0x0;
                FUN_0804fc9a();
                FUN_0804fc9a();
                FUN_0804fc9a();
                FUN_0804edbe();
                FUN_0804edbe();
                FUN_0804edbe();
                local_c0 = (BigNum *)0x0;
                FUN_08051856();
                FUN_0804edbe();
                FUN_0805091c();
                FUN_0805091c();
                FUN_0805091c();
                FUN_0804f970();
                ppBVar27 = &local_12c;
                goto LAB_080614ac;
              }
              FUN_0804edbe();
            }
          }
        }
        ppBVar32 = &local_12c;
        pBStack_130 = (BigNum *)0x806118d;
        FUN_0805091c();
        pBStack_130 = (BigNum *)0x8061198;
        FUN_0805091c();
        pBStack_130 = (BigNum *)0x80611a3;
        FUN_0805091c();
        pBStack_130 = (BigNum *)0x80611ae;
        FUN_0804f970();
        pBVar36 = extraout_ECX_04;
        goto LAB_080615ed;
      }
      if (*(int *)(param_1 + 0x88) == 2) {
        local_bc = local_104;
        pBStack_130 = (BigNum *)0x80611df;
        local_b8 = local_110;
        iVar21 = FUN_0804f62c();
        local_94 = (BigNum *)0x0;
        local_90 = (BigNum *)0x0;
        ppBVar32 = &local_12c;
        if ((iVar21 == 3) && (ppBVar32 = &local_12c, 1 < (uint)((int)local_110 - (int)local_bc))) {
          sVar11 = *(short *)local_bc;
          pBStack_130 = (BigNum *)0x806122c;
          FUN_0804f662();
          pBStack_130 = (BigNum *)0x806123d;
          cVar7 = FUN_0805e85c();
          ppBVar32 = &local_12c;
          if (cVar7 != '\0') {
            sVar10 = CONCAT11((char)sVar11,(char)((ushort)sVar11 >> 8));
            bVar40 = 2 < (ushort)(sVar10 - 0x17U);
            local_10c = (int *)CONCAT31(local_10c._1_3_,bVar40 && sVar11 != 0x1d00);
            ppBVar32 = &local_12c;
            if (!bVar40 || sVar11 == 0x1d00) {
              local_114 = local_bc;
              local_110 = local_b8;
              if (sVar11 == 0x1d00) {
                puVar16 = malloc(0xc);
                *puVar16 = &DAT_08064640;
                FUN_080505f8();
                *puVar16 = 0x8066edc;
                puVar16[1] = 0;
                puVar16[2] = 0;
                FUN_0804ff50();
                FUN_08050a6a();
                FUN_0805091c();
                FUN_08051856();
                ppBVar27 = &local_12c;
              }
              else {
                uVar35 = sVar10 - 0x13;
                local_9c = malloc(0x38);
                pBVar36 = local_9c;
                for (iVar21 = 0xe; iVar21 != 0; iVar21 = iVar21 + -1) {
                  *(undefined4 *)pBVar36 = 0;
                  pBVar36 = pBVar36 + (uint)bVar41 * -8 + 4;
                }
                *(undefined **)local_9c = &DAT_08064640;
                pBVar36 = (BigNum *)0x0;
                FUN_080505f8();
                *(undefined4 *)local_9c = 0x8066ad0;
                local_12c = local_9c + 4;
                FUN_08050d86();
                FUN_08050d86();
                FUN_08050d68();
                local_120 = local_9c + 0x30;
                FUN_08050d68();
                ECPoint::fromOctetString();
                ppBVar27 = &pBStack_130;
                ppBVar32 = &pBStack_130;
                if (uVar35 < 7) {
                  pBVar36 = (BigNum *)(int)(char)(&DAT_080650e4)[uVar35];
                }
                FUN_08050d86();
                cVar8 = FUN_0804fe56();
                cVar7 = '\0';
                if (cVar8 != '\0') {
                  cVar7 = FUN_0804fe56();
                }
                FUN_0804f6cc();
                if (cVar7 == '\0') {
                  cVar7 = FUN_080593c0();
                  local_10c = (int *)CONCAT31(local_10c._1_3_,cVar7);
                  if (cVar7 != '\0') {
                    pBVar36 = (BigNum *)&local_80;
                    cVar7 = BigNum::operator<(pBVar36,local_120);
                    local_10c = (int *)CONCAT31(local_10c._1_3_,cVar7);
                    if (cVar7 != '\0') {
                      cVar7 = BigNum::operator<((BigNum *)local_78,local_120);
                      local_10c = (int *)CONCAT31(local_10c._1_3_,cVar7);
                      if (cVar7 != '\0') {
                        FUN_0805b76a();
                        cVar7 = FUN_0805ef2a();
                        local_10c = (int *)CONCAT31(local_10c._1_3_,cVar7);
                        if (cVar7 != '\0') {
                          FUN_0804fc9a();
                          FUN_0804fc9a();
                        }
                        FUN_0805931c();
                      }
                    }
                  }
                }
                FUN_0804f6cc();
                if ((char)local_10c == '\0') {
                  FUN_0805997c();
                  goto LAB_08061496;
                }
                local_9c = (BigNum *)0x0;
                FUN_08051856();
                FUN_0805997c();
              }
              *(undefined4 *)((int)ppBVar27 + 0xfffffffc) = 0x8061494;
              FUN_0805091c();
              goto LAB_080614ac;
            }
          }
        }
LAB_08061496:
        *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x80614a1;
        FUN_0805091c();
        pBVar36 = extraout_ECX_05;
      }
      else {
        local_114 = local_104;
LAB_080614ac:
        *(BigNum **)((int)ppBVar27 + 0xfffffffc) = pBVar36;
        *(BigNum **)((int)ppBVar27 + 0xfffffff8) = pBVar36;
        *(char **)((int)ppBVar27 + 0xfffffff4) = "srvparams bytes: ";
        *(undefined1 **)((int)ppBVar27 + 0xfffffff0) = cout;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80614bd;
        uVar19 = operator<<(*(ostream **)((int)ppBVar27 + 0xfffffff0),
                            *(char **)((int)ppBVar27 + 0xfffffff4));
        *(int *)((int)ppBVar27 + 0xfffffff4) = (int)local_114 - (int)local_104;
        *(undefined4 *)((int)ppBVar27 + 0xfffffff0) = uVar19;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80614d2;
        uVar19 = operator<<(*(ostream **)((int)ppBVar27 + 0xfffffff0),
                            *(uint *)((int)ppBVar27 + 0xfffffff4));
        *(char **)((int)ppBVar27 + 0xfffffff4) = ", signature bytes: ";
        *(undefined4 *)((int)ppBVar27 + 0xfffffff0) = uVar19;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80614df;
        uVar19 = operator<<(*(ostream **)((int)ppBVar27 + 0xfffffff0),
                            *(char **)((int)ppBVar27 + 0xfffffff4));
        *(int *)((int)ppBVar27 + 0xfffffff4) = (int)local_110 - (int)local_114;
        *(undefined4 *)((int)ppBVar27 + 0xfffffff0) = uVar19;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80614f4;
        uVar19 = operator<<(*(ostream **)((int)ppBVar27 + 0xfffffff0),
                            *(uint *)((int)ppBVar27 + 0xfffffff4));
        *(undefined4 *)((int)ppBVar27 + 0xfffffff0) = uVar19;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80614fc;
        endl(*(ostream **)((int)ppBVar27 + 0xfffffff0));
        if (local_128 == 1) {
LAB_08061665:
          local_8c = (BigNum *)0x0;
          *(undefined4 *)((int)ppBVar27 + 0xfffffffc) = 0x8061680;
          FUN_08051856();
          *(undefined4 *)(param_1 + 0x158) = 3;
          *(undefined4 *)((int)ppBVar27 + 0xfffffffc) = 0x8061698;
          FUN_0805b8f6();
          return (BigNum *)0x1;
        }
        iVar21 = *(int *)(local_11c + *(int *)(*(int *)local_11c + -0x28));
        *(BigNum **)((int)ppBVar27 + 0xfffffff0) = local_11c + *(int *)(*(int *)local_11c + -0x28);
        pcVar4 = *(code **)(iVar21 + 0x18);
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x8061526;
        (*pcVar4)();
        local_bc = local_114;
        local_b8 = local_110;
        *(BigNum ***)((int)ppBVar27 + 0xfffffff4) = &local_88;
        *(BigNum ***)((int)ppBVar27 + 0xfffffff0) = &local_80;
        *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x806155f;
        cVar7 = FUN_0805e966();
        pBVar36 = extraout_ECX_06;
        ppBVar32 = ppBVar27;
        if (cVar7 != '\0') {
          local_70 = (undefined1  [4])0x0;
          local_6c = (BigNum *)0x0;
          *(undefined4 *)((int)ppBVar27 + 0xfffffffc) = *(undefined4 *)(param_1 + 0x38);
          *(undefined4 *)((int)ppBVar27 + 0xfffffff8) = *(undefined4 *)(param_1 + 0x34);
          *(undefined4 *)((int)ppBVar27 + 0xfffffff4) = 0;
          *(undefined1 **)((int)ppBVar27 + 0xfffffff0) = local_70;
          *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x8061589;
          vector<unsigned_char>::insert
                    (*(uchar **)((int)ppBVar27 + 0xfffffff0),*(uchar **)((int)ppBVar27 + 0xfffffff4)
                     ,*(uchar **)((int)ppBVar27 + 0xfffffff8));
          *(undefined4 *)((int)ppBVar27 + 0xffffffec) = *(undefined4 *)(param_1 + 0x40);
          *(undefined4 *)((int)ppBVar27 + 0xffffffe8) = *(undefined4 *)(param_1 + 0x3c);
          *(BigNum **)((int)ppBVar27 + 0xffffffe4) = local_6c;
          *(undefined1 **)((int)ppBVar27 + 0xffffffe0) = local_70;
          *(undefined4 *)((int)ppBVar27 + 0xffffffdc) = 0x806159e;
          vector<unsigned_char>::insert
                    (*(uchar **)((int)ppBVar27 + 0xffffffe0),*(uchar **)((int)ppBVar27 + 0xffffffe4)
                     ,*(uchar **)((int)ppBVar27 + 0xffffffe8));
          *(BigNum **)((int)ppBVar27 + 0xfffffffc) = local_114;
          *(BigNum **)((int)ppBVar27 + 0xfffffff8) = local_104;
          *(BigNum **)((int)ppBVar27 + 0xfffffff4) = local_6c;
          *(undefined1 **)((int)ppBVar27 + 0xfffffff0) = local_70;
          *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80615b6;
          vector<unsigned_char>::insert
                    (*(uchar **)((int)ppBVar27 + 0xfffffff0),*(uchar **)((int)ppBVar27 + 0xfffffff4)
                     ,*(uchar **)((int)ppBVar27 + 0xfffffff8));
          *(undefined4 *)((int)ppBVar27 + 0xfffffff8) = local_84;
          *(BigNum **)((int)ppBVar27 + 0xfffffff4) = local_88;
          *(BigNum **)((int)ppBVar27 + 0xfffffff0) = local_80;
          *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80615db;
          cVar7 = FUN_0805f0e0();
          *(undefined4 *)((int)ppBVar27 + 0xffffffec) = 0x80615e4;
          FUN_0805091c();
          pBVar36 = extraout_ECX_07;
          if (cVar7 != '\0') goto LAB_08061665;
        }
      }
LAB_080615ed:
      *(BigNum **)((int)ppBVar32 + 0xfffffffc) = pBVar36;
      *(BigNum **)((int)ppBVar32 + 0xfffffff8) = pBVar36;
      *(undefined4 *)((int)ppBVar32 + 0xfffffff4) = 3;
      *(undefined **)((int)ppBVar32 + 0xfffffff0) = &DAT_0806400e;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8061608;
      FUN_080589fe();
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x806160f;
      uVar19 = FUN_0804f610();
      *(char **)((int)ppBVar32 + 0xfffffff4) = " parsing ServerKeyExchange";
      *(undefined4 *)((int)ppBVar32 + 0xfffffff0) = uVar19;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x806161f;
      uVar19 = operator<<(*(ostream **)((int)ppBVar32 + 0xfffffff0),
                          *(char **)((int)ppBVar32 + 0xfffffff4));
      *(undefined4 *)((int)ppBVar32 + 0xfffffff0) = uVar19;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8061627;
      endl(*(ostream **)((int)ppBVar32 + 0xfffffff0));
      *(undefined1 **)((int)ppBVar32 + 0xfffffff0) = local_70;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x806162f;
      string::freeptr();
      local_70 = (undefined1  [4])0x11;
      *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_6c;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x806163e;
      string::string(*(string **)((int)ppBVar32 + 0xfffffff0));
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x806164d;
      FUN_0805a4bc();
      *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_6c;
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8061655;
      string::freeptr();
      *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8061660;
      FUN_0805b8f6();
      return (BigNum *)0x0;
    case (BigNum)0xd:
      puVar17 = malloc(0x18);
      puVar16 = puVar17;
      for (iVar21 = 6; iVar21 != 0; iVar21 = iVar21 + -1) {
        *puVar16 = 0;
        puVar16 = puVar16 + (uint)bVar41 * -2 + 1;
      }
      pvVar5 = *(void **)(param_1 + 0x160);
      *(undefined4 **)(param_1 + 0x160) = puVar17;
      if (pvVar5 != (void *)0x0) {
        pBStack_130 = (BigNum *)0x80616d1;
        FUN_080520f4();
        free(pvVar5);
      }
      local_110 = local_104 + (int)local_10c;
      if ((local_110 != local_104) &&
         (pBVar36 = (BigNum *)(uint)(byte)*local_104, pBVar36 <= (BigNum *)((int)local_10c + -1))) {
        pBVar15 = local_104 + 1;
        local_10c = *(int **)(param_1 + 0x160);
        local_114 = (BigNum *)CONCAT22(local_114._2_2_,(short)*(undefined4 *)(param_1 + 0x80));
        pBVar25 = pBVar36;
        pBVar38 = pBVar15;
        for (; pBVar36 != (BigNum *)0x0; pBVar36 = pBVar36 + -1) {
          BVar1 = *pBVar38;
          pBVar38 = pBVar38 + 1;
          local_70[0] = BVar1;
          pBStack_130 = (BigNum *)0x8061754;
          local_104 = pBVar25;
          FUN_08059a02();
          pBVar25 = local_104;
        }
        pBVar15 = pBVar15 + (int)pBVar25;
        param_2 = (BigNum *)0x0;
        if (((short)local_114 == 0x303) || ((short)local_114 == -0x103)) {
          if (1 < (uint)((int)local_110 - (int)pBVar15)) {
            uVar35 = *(ushort *)pBVar15;
            local_104 = pBVar15 + 2;
            uVar23 = CONCAT11((char)uVar35,(char)(uVar35 >> 8));
            if (((uint)uVar23 <= (uint)((int)local_110 - (int)local_104)) && ((uVar35 & 0x100) == 0)
               ) {
              local_114 = (BigNum *)(local_10c + 2);
              pBStack_130 = (BigNum *)0x80617d6;
              FUN_080506a4();
              uVar24 = (uint)(uVar23 >> 1);
              pBVar25 = extraout_ECX_08;
              pBVar36 = local_104;
              for (uVar20 = uVar24; uVar20 != 0; uVar20 = uVar20 - 1) {
                uVar2 = *(undefined2 *)pBVar36;
                pBVar36 = pBVar36 + 2;
                local_70._0_2_ = uVar2;
                pBStack_130 = (BigNum *)0x80617ff;
                FUN_0805084a();
                pBVar25 = extraout_ECX_09;
              }
              pBVar15 = local_104 + uVar24 * 2;
              goto LAB_0806180b;
            }
          }
        }
        else {
LAB_0806180b:
          param_2 = (BigNum *)0x0;
          if (1 < (uint)((int)local_110 - (int)pBVar15)) {
            pBVar36 = pBVar15 + 2;
            uVar20 = (uint)CONCAT11((char)*(undefined2 *)pBVar15,
                                    (char)((ushort)*(undefined2 *)pBVar15 >> 8));
            if (uVar20 == (int)local_110 - (int)pBVar36) {
              while( true ) {
                if (uVar20 == 0) {
                  pBStack_130 = pBVar25;
                  FUN_080589fe();
                  poVar13 = (ostream *)FUN_0804f610();
                  poVar13 = (ostream *)operator<<(poVar13," rx ");
                  puVar16 = *(undefined4 **)(param_1 + 0x160);
                  operator<<(poVar13,"(CertificateReq) certTypes: ");
                  pBStack_130 = (BigNum *)puVar16[1];
                  local_104 = pBStack_130;
                  for (pBVar36 = (BigNum *)*puVar16; local_104 != pBVar36; pBVar36 = pBVar36 + 1) {
                    poVar18 = (ostream *)operator<<(poVar13,(uint)(byte)*pBVar36);
                    pBStack_130 = (BigNum *)operator<<(poVar18,", ");
                  }
                  pBStack_130 = pBVar36;
                  operator<<(poVar13," distingNames: ");
                  iVar21 = puVar16[5];
                  piVar14 = (int *)&stack0xfffffec4;
                  local_104 = (BigNum *)iVar21;
                  for (iVar39 = puVar16[4]; local_104 != (BigNum *)iVar39; iVar39 = iVar39 + 0x24) {
                    piVar14[3] = iVar21;
                    piVar14[2] = iVar21;
                    piVar14[1] = (int)&DAT_080638d4;
                    *piVar14 = (int)poVar13;
                    piVar14[-1] = 0x8061a48;
                    local_10c = (int *)operator<<((ostream *)*piVar14,(char *)piVar14[1]);
                    piVar14[1] = iVar39;
                    *piVar14 = (int)local_70;
                    piVar14[-1] = 0x8061a5a;
                    DistinguishedName::str();
                    piVar14[2] = 0x8061a6a;
                    uVar19 = FUN_0804f275();
                    piVar14[2] = extraout_ECX_11;
                    piVar14[1] = extraout_ECX_11;
                    *piVar14 = (int)&DAT_080643a8;
                    piVar14[-1] = uVar19;
                    piVar14[-2] = 0x8061a77;
                    operator<<((ostream *)piVar14[-1],(char *)*piVar14);
                    piVar14[-1] = (int)local_70;
                    piVar14[-2] = 0x8061a7f;
                    iVar21 = string::freeptr();
                    piVar14 = piVar14 + -1;
                  }
                  *piVar14 = (int)poVar13;
                  piVar14[-1] = 0x8061a8a;
                  endl((ostream *)*piVar14);
                  *piVar14 = (int)&local_80;
                  piVar14[-1] = 0x8061a95;
                  string::freeptr();
                  *(undefined4 *)(param_1 + 0x158) = 4;
                  return (BigNum *)0x1;
                }
                if (uVar20 == 1) break;
                local_110 = (BigNum *)(uVar20 - 2);
                local_104 = (BigNum *)
                            (uint)CONCAT11((char)*(undefined2 *)pBVar36,
                                           (char)((ushort)*(undefined2 *)pBVar36 >> 8));
                if (local_110 < local_104) break;
                pBVar15 = pBVar36 + 2;
                puVar16 = (undefined4 *)local_70;
                for (iVar21 = 9; iVar21 != 0; iVar21 = iVar21 + -1) {
                  *puVar16 = 0;
                  puVar16 = puVar16 + (uint)bVar41 * -2 + 1;
                }
                pBStack_130 = (BigNum *)0x806187e;
                FUN_08050352();
                pBStack_130 = (BigNum *)0x806188f;
                FUN_0805ffd2();
                pBStack_130 = (BigNum *)0x8061897;
                FUN_08050b38();
                local_60 = 0;
                pBVar36 = pBVar15 + (int)local_104;
                local_64 = (BigNum *)0x0;
                local_5c = 0;
                local_70 = (undefined1  [4])0xffffffff;
                pBStack_130 = (BigNum *)0x80618c0;
                local_6c = pBVar15;
                local_68 = pBVar36;
                FUN_0804f76a();
                uVar19 = local_5c;
                local_120 = local_64;
                local_114 = (BigNum *)local_10c[5];
                param_2 = local_114 + -0x24;
                pBStack_130 = (BigNum *)0x80618e6;
                cVar7 = FUN_0804f76a();
                local_11c = local_120;
                if (cVar7 == '\0') {
                  local_11c = (BigNum *)0x0;
                  uVar19 = 0;
                }
                pBStack_130 = (BigNum *)0x0;
                local_118 = uVar19;
                cVar7 = DistinguishedName::parse();
                if (cVar7 == '\0') break;
                uVar20 = (int)local_110 - (int)local_104;
                pBVar25 = extraout_ECX_10;
              }
            }
          }
        }
      }
      pBStack_130 = param_2;
      FUN_080589fe();
      poVar13 = (ostream *)FUN_0804f610();
      poVar13 = (ostream *)operator<<(poVar13," parsing CertificateRequest");
      endl(poVar13);
      string::freeptr();
      local_70 = (undefined1  [4])0x11;
      string::string((string *)&local_6c);
LAB_080619de:
      piVar26 = (int *)&stack0xfffffec4;
      FUN_0805a4bc();
      goto LAB_0806017c;
    case (BigNum)0xe:
      local_a4 = (BigNum *)0x0;
      local_a0 = (BigNum *)0x0;
      local_ac = 0;
      if (*(int *)(param_1 + 0x160) == 0) {
LAB_08061bdc:
        puVar29 = (undefined1 *)&local_12c;
        local_9c = (BigNum *)0x0;
        local_98 = (BigNum *)0x0;
        uVar19 = *(undefined4 *)(param_1 + 0xa0);
        local_10c = *(int **)(*(int *)(param_1 + 0x30) + 4);
        if (*(int *)(param_1 + 0x88) == 0) {
          local_104 = *(BigNum **)(param_1 + 0x164);
        }
        else {
          local_104 = *(BigNum **)(param_1 + 0x168);
        }
        iVar21 = (**(code **)(*(int *)(local_104 + *(int *)(*(int *)local_104 + -0x28)) + 0x18))();
        if (iVar21 == 2) {
          pBStack_130 = (BigNum *)0x8061c53;
          FUN_08050d9a();
          pBStack_130 = (BigNum *)0x8061c6a;
          FUN_0805b810();
          pBStack_130 = (BigNum *)0x8061c7d;
          FUN_0805efca();
          pBStack_130 = (BigNum *)0x8061c8a;
          FUN_08050a6a();
          pBStack_130 = (BigNum *)0x8061c91;
          FUN_0805091c();
          if (local_a0 != local_a4) {
            pBStack_130 = local_a4;
            BigNum::bin((uint)&local_80);
            ECPoint::toBin((uint)&local_8c,SUB41(&local_6c,0));
            FUN_0805091c();
            FUN_08058f1a();
            *local_9c = (BigNum)((char)local_88 - (char)local_8c);
            iVar21 = (int)local_88 - (int)local_8c;
            pBVar36 = local_9c + 1;
            for (; iVar21 != 0; iVar21 = iVar21 + -1) {
              *pBVar36 = *local_8c;
              local_8c = local_8c + (uint)bVar41 * -2 + 1;
              pBVar36 = pBVar36 + (uint)bVar41 * -2 + 1;
            }
            FUN_0805091c();
            local_70 = (undefined1  [4])0x8066d48;
            local_38 = 0x8066d94;
            FUN_0804fbdf();
            puVar30 = &stack0xfffffebc;
            FUN_0804edbe();
            FUN_0804f960();
            local_70 = (undefined1  [4])0;
            *(undefined4 *)(local_70 + iRamffffffd8) = 0;
            FUN_0804edbe();
            FUN_0804edbe();
            FUN_0804f6cc();
            FUN_0804f6cc();
LAB_08061fc0:
            *(undefined4 *)(puVar30 + -4) = 0x8061fc8;
            FUN_0804f898();
LAB_080620ee:
            puVar29 = puVar30 + 0x10;
LAB_0806236f:
            uVar35 = *(ushort *)(param_1 + 0x120);
            *(ushort *)(param_1 + 0x120) = uVar35 + 1;
            *(undefined4 *)(puVar29 + -4) = 0x806238d;
            uVar20 = FUN_080587fa();
            *(uint *)(puVar29 + -4) = (uint)uVar35;
            *(uint *)(puVar29 + -8) = uVar20 & 0xff;
            *(BigNum **)(puVar29 + -0xc) = local_98;
            *(BigNum **)(puVar29 + -0x10) = local_9c;
            *(undefined4 *)(puVar29 + -0x14) = 0x80623ad;
            FUN_0805e924();
            *(uint *)(puVar29 + -0x10) = (uint)*(ushort *)(param_1 + 0x80);
            *(undefined4 *)(puVar29 + -0x14) = 0x80623c8;
            FUN_0805a416();
            local_104 = (BigNum *)(param_1 + 0x50);
            *(undefined4 *)(puVar29 + -0x14) = 0x80623dc;
            FUN_080594e2();
            *(undefined4 *)(puVar29 + -0x14) = 0x80623e4;
            FUN_0805091c();
            *(undefined4 *)(puVar29 + -0x14) = 0x80623ef;
            FUN_0805091c();
            if (*(int *)(param_1 + 0x160) != 0) {
              iVar21 = *(int *)(param_1 + 0x30);
              *(undefined4 *)(puVar29 + -4) = 0x806240d;
              cVar7 = FUN_08050cfe();
              iVar39 = local_ac;
              if (cVar7 != '\0') {
                local_80 = (BigNum *)0x0;
                local_10c = *(int **)(iVar21 + 4);
                local_7c = 0;
                local_11c = *(BigNum **)(iVar21 + 0x20);
                local_110 = *(BigNum **)(param_1 + 0x58);
                local_114 = *(BigNum **)(param_1 + 0x5c);
                uVar35 = *(ushort *)(param_1 + 0x120);
                *(ushort *)(param_1 + 0x120) = uVar35 + 1;
                *(undefined4 *)(puVar29 + -4) = 0x8062474;
                uVar9 = FUN_080587fa();
                local_120 = (BigNum *)CONCAT31(local_120._1_3_,uVar9);
                local_70 = (undefined1  [4])0x0;
                local_6c = (BigNum *)0x0;
                *(BigNum **)(puVar29 + -4) = local_114;
                *(BigNum **)(puVar29 + -8) = local_110;
                *(int *)(puVar29 + -0xc) = iVar39;
                *(BigNum **)(puVar29 + -0x10) = local_11c;
                *(undefined4 *)(puVar29 + -0x14) = 0x80624a8;
                FUN_0805fa50();
                *(uint *)(puVar29 + -0x14) = (uint)uVar35;
                *(uint *)(puVar29 + -0x18) = (uint)local_120 & 0xff;
                *(BigNum **)(puVar29 + -0x1c) = local_6c;
                *(undefined1 (*) [4])(puVar29 + -0x20) = local_70;
                *(undefined4 *)(puVar29 + -0x24) = 0x80624c9;
                FUN_0805e924();
                *(undefined4 *)(puVar29 + -4) = 0x80624d4;
                FUN_0805091c();
                *(uint *)(puVar29 + -0x10) = (uint)*(ushort *)(param_1 + 0x80);
                *(undefined4 *)(puVar29 + -0x14) = 0x80624f6;
                FUN_0805a416();
                *(undefined4 *)(puVar29 + -0x14) = 0x8062507;
                FUN_080594e2();
                *(undefined4 *)(puVar29 + -0x14) = 0x8062512;
                FUN_0805091c();
                *(undefined4 *)(puVar29 + -0x14) = 0x806251a;
                FUN_0805091c();
              }
            }
            *(undefined4 *)(puVar29 + -4) = 0x806252b;
            FUN_080596a2();
            *(undefined4 *)(puVar29 + -4) = 0x806253a;
            FUN_08058c6a();
            *(undefined1 *)(param_1 + 0x65) = 1;
            *(undefined4 *)(puVar29 + -4) = 0x806254c;
            FUN_08059a90();
            uVar19 = *(undefined4 *)(param_1 + 0xa4);
            *(undefined1 **)(puVar29 + -4) = local_70;
            *(undefined4 *)(puVar29 + -8) = 0xf;
            *(char **)(puVar29 + -0xc) = "client finished";
            *(undefined4 *)(puVar29 + -0x10) = uVar19;
            *(undefined4 *)(puVar29 + -0x14) = 0x806257c;
            FUN_080594ee();
            local_10c = *(int **)(param_1 + 0x120);
            *(short *)(param_1 + 0x120) = (short)local_10c + 1;
            *(undefined4 *)(puVar29 + -0x14) = 0x80625a2;
            uVar20 = FUN_080587fa();
            *(uint *)(puVar29 + -0x14) = (uint)local_10c & 0xffff;
            *(uint *)(puVar29 + -0x18) = uVar20 & 0xff;
            *(BigNum ***)(puVar29 + -0x1c) = &local_64;
            *(undefined1 **)(puVar29 + -0x20) = local_70;
            *(undefined4 *)(puVar29 + -0x24) = 0x80625bd;
            FUN_0805e924();
            *(undefined4 *)(puVar29 + -4) = 0x80625ce;
            FUN_080594e2();
            puVar34 = puVar29 + -0x10;
            *(uint *)(puVar29 + -0x10) = (uint)*(ushort *)(param_1 + 0x80);
            *(undefined4 *)(puVar29 + -0x14) = 0x80625ea;
            FUN_0805a416();
            *(undefined4 *)(puVar29 + -0x14) = 0x80625f2;
            FUN_08059e70();
            *(undefined4 *)(puVar29 + -0x14) = 0x80625fa;
            FUN_0805091c();
            *(undefined4 *)(param_1 + 0x158) = 5;
LAB_0806260d:
            *(undefined4 *)(puVar34 + -4) = 0x8062612;
            FUN_0805091c();
            return (BigNum *)0x1;
          }
          param_2 = (BigNum *)&local_34;
          local_70 = (undefined1  [4])0x8066d48;
          local_38 = 0x8066d94;
          pBStack_130 = (BigNum *)0x8061cbd;
          FUN_0804fbdf();
          puVar28 = &stack0xfffffec4;
          FUN_0804edbe();
          FUN_0804f960();
          local_70 = (undefined1  [4])0;
          *(undefined4 *)(local_70 + iRamffffffd8) = 0;
          FUN_0804edbe();
          FUN_0804edbe();
          FUN_0804f6cc();
          FUN_0804f6cc();
LAB_08061f2b:
          *(undefined4 *)(puVar28 + -4) = 0x8061f33;
          FUN_0804f898();
LAB_0806208b:
          puVar31 = puVar28 + 0x10;
        }
        else {
          iVar21 = (**(code **)(*(int *)(local_104 + *(int *)(*(int *)local_104 + -0x28)) + 0x18))()
          ;
          if (iVar21 == 4) {
            local_70 = (undefined1  [4])&DAT_08064640;
            param_2 = (BigNum *)&local_80;
            pBStack_130 = (BigNum *)0x8061e47;
            FUN_080505f8();
            *(undefined4 *)(local_70 + iRamffffffd8) = 0;
            local_6c = (BigNum *)0x0;
            local_68 = (BigNum *)0x0;
            pBStack_130 = (BigNum *)0x8061e6f;
            FUN_08050606();
            local_70 = (undefined1  [4])0x8066c5c;
            local_60 = 0;
            local_5c = 0;
            local_64 = (BigNum *)0x8066ca8;
            pBStack_130 = (BigNum *)0x0;
            X25519PrivateKey::generatePrivKey((Rng *)local_70);
            X25519PrivateKey::exchangeKeys(param_2,local_70);
            FUN_08050a6a();
            FUN_0805091c();
            if ((int)local_a0 - (int)local_a4 == 0x20) {
              FUN_08058f1a();
              *local_9c = (BigNum)((char)local_68 - (char)local_6c);
              pBVar36 = local_6c;
              pBVar15 = local_9c + 1;
              for (iVar21 = (int)local_68 - (int)local_6c; iVar21 != 0; iVar21 = iVar21 + -1) {
                *pBVar15 = *pBVar36;
                pBVar36 = pBVar36 + (uint)bVar41 * -2 + 1;
                pBVar15 = pBVar15 + (uint)bVar41 * -2 + 1;
              }
              local_70 = (undefined1  [4])0x8066c5c;
              local_64 = (BigNum *)(X25519PrivateKey::vtable + 0x74);
              puVar30 = &stack0xfffffec0;
              (*(code *)PTR_memset_080673ac)();
              FUN_0805091c();
              FUN_0804f960();
              local_70 = (undefined1  [4])0;
              *(undefined4 *)(local_70 + iRamffffffd8) = 0;
              FUN_0805091c();
              goto LAB_08061fc0;
            }
            local_70 = (undefined1  [4])0x8066c5c;
            local_64 = (BigNum *)(X25519PrivateKey::vtable + 0x74);
            puVar28 = &stack0xfffffec0;
            (*(code *)PTR_memset_080673ac)();
            FUN_0805091c();
            FUN_0804f960();
            local_70 = (undefined1  [4])0;
            *(undefined4 *)(local_70 + iRamffffffd8) = 0;
            FUN_0805091c();
            goto LAB_08061f2b;
          }
          iVar21 = (**(code **)(*(int *)(local_104 + *(int *)(*(int *)local_104 + -0x28)) + 0x18))()
          ;
          if (iVar21 == 1) {
            param_2 = (BigNum *)&local_8c;
            pBStack_130 = (BigNum *)0x8061fff;
            FUN_0805b7a8();
            pBStack_130 = (BigNum *)0x8062012;
            FUN_08050e36();
            pBStack_130 = (BigNum *)0x8062022;
            FUN_08050e36();
            puVar28 = &stack0xfffffec4;
            FUN_0805b6ba();
            FUN_0804edbe();
            FUN_0804edbe();
            FUN_0805f7da();
            FUN_08050a6a();
            FUN_0805091c();
            if (local_a0 != local_a4) {
              pBStack_130 = local_a4;
              BigNum::bin((uint)&local_80);
              FUN_08058f1a();
              *(ushort *)local_9c =
                   CONCAT11((char)(local_7c - (int)local_80),
                            (char)((uint)(local_7c - (int)local_80) >> 8));
              pBVar36 = local_80;
              pBVar15 = local_9c + 2;
              for (iVar21 = local_7c - (int)local_80; iVar21 != 0; iVar21 = iVar21 + -1) {
                *pBVar15 = *pBVar36;
                pBVar36 = pBVar36 + (uint)bVar41 * -2 + 1;
                pBVar15 = pBVar15 + (uint)bVar41 * -2 + 1;
              }
              FUN_0805091c();
              puVar30 = &stack0xfffffec0;
              FUN_0804fbfc();
              goto LAB_080620ee;
            }
            FUN_0804fbfc();
            goto LAB_0806208b;
          }
          iVar21 = (**(code **)(*(int *)(local_104 + *(int *)(*(int *)local_104 + -0x28)) + 0x18))()
          ;
          puVar31 = (undefined1 *)&local_12c;
          if (iVar21 == 0) {
            pBStack_130 = (BigNum *)0x806212d;
            pBStack_130 = (BigNum *)FUN_08058f1a();
            (**(code **)(*local_10c + 8))();
            pBVar36 = local_a0;
            param_2 = local_a4;
            *(ushort *)local_a4 = CONCAT11((char)uVar19,(char)((uint)uVar19 >> 8));
            uVar20 = FUN_0805e7e0();
            if ((uVar20 < 0xb) || (uVar24 = (int)pBVar36 - (int)param_2, uVar20 - 0xb < uVar24)) {
              local_94 = (BigNum *)0x0;
              local_90 = (BigNum *)0x0;
            }
            else {
              pBStack_130 = (BigNum *)0x8062187;
              FUN_0805002a();
              uVar20 = ~uVar24;
              *local_8c = (BigNum)0x0;
              local_8c[1] = (BigNum)0x2;
              pBVar36 = local_88 + ((int)param_2 - (int)pBVar36);
              for (; uVar24 != 0; uVar24 = uVar24 - 1) {
                *pBVar36 = *param_2;
                param_2 = param_2 + (uint)bVar41 * -2 + 1;
                pBVar36 = pBVar36 + (uint)bVar41 * -2 + 1;
              }
              pBVar25 = local_88 + uVar20;
              pBVar36 = local_8c + 2;
              pBStack_130 = local_8c;
              pBStack_130 = (BigNum *)(**(code **)(*local_10c + 8))();
              pBVar15 = extraout_ECX_12;
              for (; pBVar25 != pBVar36; pBVar36 = pBVar36 + 1) {
                while (*pBVar36 == (BigNum)0x0) {
                  pBStack_130 = (BigNum *)(**(code **)(*local_10c + 8))();
                  pBVar15 = extraout_ECX_13;
                }
              }
              *pBVar25 = (BigNum)0x0;
              param_2 = (BigNum *)local_70;
              pBStack_130 = pBVar15;
              BigNum::BigNum();
              expByMod(param_2,(BigNum *)&local_80,local_104 + 0xc);
              FUN_0805e7e0();
              BigNum::bin((uint)&local_94);
              FUN_0804edbe();
              FUN_0804edbe();
              FUN_0805091c();
              puVar29 = &stack0xfffffecc;
            }
            if (local_90 != local_94) {
              *(undefined4 *)(puVar29 + 0xfffffffc) = 0x80622ae;
              FUN_08058f1a();
              *(ushort *)local_9c =
                   CONCAT11((char)((int)local_90 - (int)local_94),
                            (char)((uint)((int)local_90 - (int)local_94) >> 8));
              pBVar36 = local_94;
              pBVar15 = local_9c + 2;
              for (iVar21 = (int)local_90 - (int)local_94; iVar21 != 0; iVar21 = iVar21 + -1) {
                *pBVar15 = *pBVar36;
                pBVar36 = pBVar36 + (uint)bVar41 * -2 + 1;
                pBVar15 = pBVar15 + (uint)bVar41 * -2 + 1;
              }
              *(undefined4 *)(puVar29 + 0xfffffffc) = 0x80622e9;
              FUN_0805091c();
              goto LAB_0806236f;
            }
            *(undefined4 *)(puVar29 + 0xfffffffc) = 0x806229c;
            FUN_0805091c();
            puVar31 = puVar29;
          }
        }
        *(BigNum **)(puVar31 + 0xfffffffc) = param_2;
        *(BigNum **)(puVar31 + 0xfffffff8) = param_2;
        *(undefined4 *)(puVar31 + 0xfffffff4) = 3;
        puVar33 = puVar31 + 0xfffffff0;
        *(undefined **)(puVar31 + 0xfffffff0) = &DAT_0806400e;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x806230c;
        FUN_080589fe();
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062313;
        uVar19 = FUN_0804f610();
        *(char **)(puVar31 + 0xfffffff4) = " exchanging keys";
        *(undefined4 *)(puVar31 + 0xfffffff0) = uVar19;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062320;
        uVar19 = operator<<(*(ostream **)(puVar31 + 0xfffffff0),*(char **)(puVar31 + 0xfffffff4));
        *(undefined4 *)(puVar31 + 0xfffffff0) = uVar19;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062328;
        endl(*(ostream **)(puVar31 + 0xfffffff0));
        *(undefined1 **)(puVar31 + 0xfffffff0) = local_70;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062330;
        string::freeptr();
        local_70 = (undefined1  [4])0x27;
        *(BigNum ***)(puVar31 + 0xfffffff0) = &local_6c;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x806233f;
        string::string(*(string **)(puVar31 + 0xfffffff0));
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x806234e;
        FUN_0805a4bc();
        *(BigNum ***)(puVar31 + 0xfffffff0) = &local_6c;
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062356;
        string::freeptr();
        *(undefined4 *)(puVar31 + 0xffffffec) = 0x8062361;
        FUN_0805091c();
      }
      else {
        iVar21 = *(int *)(param_1 + 0x30);
        if (*(int *)(iVar21 + 0x1c) == *(int *)(iVar21 + 0x18)) {
LAB_08061b5c:
          local_70 = (undefined1  [4])0x0;
          local_6c = (BigNum *)0x0;
          *(short *)(param_1 + 0x120) = *(short *)(param_1 + 0x120) + 1;
          pBStack_130 = (BigNum *)0x8061b82;
          FUN_080587fa();
          local_104 = *(BigNum **)(param_1 + 0x30);
          pBStack_130 = param_2;
          FUN_0805edb4();
          FUN_0805a416();
          FUN_080594e2();
          FUN_0805091c();
          FUN_08050f2a();
          goto LAB_08061bdc;
        }
        pBStack_130 = (BigNum *)0x8061ae3;
        cVar7 = FUN_08050cfe();
        if (cVar7 == '\0') goto LAB_08061b5c;
        param_2 = *(BigNum **)(param_1 + 0x80);
        puVar33 = &stack0xfffffec4;
        (**(code **)(*(int *)((int)*(int **)(iVar21 + 0x20) +
                             *(int *)(**(int **)(iVar21 + 0x20) + -0x28)) + 0x18))();
        cVar7 = FUN_08058b03();
        if (cVar7 != '\0') {
          if (*(int *)(param_1 + 0x160) == 0) goto LAB_08061bdc;
          goto LAB_08061b5c;
        }
        local_70 = (undefined1  [4])0x2b;
        string::string((string *)&local_6c);
        FUN_0805a4bc();
        string::freeptr();
      }
      ppBVar32 = (BigNum **)(puVar33 + 0x10);
      break;
    case (BigNum)0x14:
      pBStack_130 = (BigNum *)0x8062628;
      local_110 = (BigNum *)&local_12c;
      cVar7 = FUN_080587fa();
      if (cVar7 != '\0') {
        pBStack_130 = (BigNum *)0x8062639;
        FUN_08050912();
        *(undefined4 *)(param_1 + 0x140) = 0;
        *(undefined4 *)(param_1 + 0x144) = 100;
        *(undefined4 *)(param_1 + 0x13c) = 1;
      }
      if (local_10c == (int *)0xc) {
        pBStack_130 = (BigNum *)local_70;
        ppBVar32 = &local_80;
        for (iVar21 = 3; iVar21 != 0; iVar21 = iVar21 + -1) {
          *ppBVar32 = *(BigNum **)local_104;
          local_104 = local_104 + (uint)bVar41 * -8 + 4;
          ppBVar32 = ppBVar32 + (uint)bVar41 * -2 + 1;
        }
        local_104 = *(BigNum **)(param_1 + 0x44);
        FUN_080594ee();
        pBVar36 = (BigNum *)FUN_08058f68();
        if ((char)pBVar36 != '\0') {
          *(undefined1 *)(param_1 + 0x6c) = 1;
          return pBVar36;
        }
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," wrong Finished hash");
        endl(poVar13);
        string::freeptr();
        local_8c = (BigNum *)0x28;
        string::string((string *)&local_88);
      }
      else {
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," parsing Finished");
        endl(poVar13);
        string::freeptr();
        local_70 = (undefined1  [4])0x11;
        string::string((string *)&local_6c);
      }
      FUN_0805a4bc();
      string::freeptr();
      return (BigNum *)0x0;
    case (BigNum)0x16:
      pBVar15 = *(BigNum **)(param_1 + 400);
      if (((char)pBVar15 == '\0') || (*(char *)(param_1 + 0x191) != '\0')) {
        pBStack_130 = pBVar36;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," unexpected CertificateStatus");
        endl(poVar13);
        string::freeptr();
        local_70 = (undefined1  [4])0x10;
        string::string((string *)&local_6c);
        goto LAB_080619de;
      }
      local_80 = (BigNum *)CONCAT31(local_80._1_3_,1);
      local_7c = 0;
      local_78[0] = 0;
      if ((local_10c < (int *)0x4) ||
         (local_80 = (BigNum *)CONCAT31(local_80._1_3_,*local_104), *local_104 != (BigNum)0x1)) {
LAB_08060ac2:
        pBStack_130 = pBVar36;
        FUN_080589fe();
        poVar13 = (ostream *)FUN_0804f610();
        poVar13 = (ostream *)operator<<(poVar13," parsing CertificateStatus");
        endl(poVar13);
        string::freeptr();
        local_70 = (undefined1  [4])0x11;
        string::string((string *)&local_6c);
        FUN_0805a4bc();
        string::freeptr();
        ppBVar32 = &local_12c;
      }
      else {
        pBStack_130 = (BigNum *)0x8060a54;
        piVar14 = (int *)FUN_080587e6();
        pBVar36 = (BigNum *)CONCAT31((int3)((uint)extraout_ECX_00 >> 8),piVar14 == (int *)0x0);
        if (local_10c + -1 != piVar14 || piVar14 == (int *)0x0) goto LAB_08060ac2;
        pBStack_130 = (BigNum *)0x8060a7f;
        FUN_080518de();
        *(undefined1 *)(param_1 + 0x191) = 1;
        local_ac = param_1;
        pBStack_130 = (BigNum *)0x8060a9c;
        FUN_0805181c();
        local_104 = local_94;
        pBStack_130 = (BigNum *)0x8060ab9;
        cVar7 = FUN_08050c9a();
        if (cVar7 != '\0') {
          ppBVar32 = &local_8c;
          for (iVar21 = 2; iVar21 != 0; iVar21 = iVar21 + -1) {
            *ppBVar32 = (BigNum *)0x0;
            ppBVar32 = ppBVar32 + (uint)bVar41 * -2 + 1;
          }
          pBVar25 = (BigNum *)(*(undefined4 **)(param_1 + 0xec))[1];
          pBStack_130 = extraout_EDX;
          for (pBVar36 = (BigNum *)**(undefined4 **)(param_1 + 0xec); pBVar25 != pBVar36;
              pBVar36 = pBVar36 + 8) {
            if ((local_104 != *(BigNum **)pBVar36) &&
               (cVar7 = DistinguishedName::operator==
                                  ((DistinguishedName *)(*(BigNum **)pBVar36 + 0x40),
                                   (DistinguishedName *)(local_104 + 0x1c)),
               pBStack_130 = extraout_EDX_00, cVar7 != '\0')) {
              pBVar25 = *(BigNum **)(pBVar36 + 4);
              auVar37 = *(undefined1 (*) [4])pBVar36;
              pBStack_130 = (BigNum *)0x8060b89;
              local_6c = pBVar25;
              FUN_08051370();
              local_8c = (BigNum *)0x0;
              pBStack_130 = (BigNum *)0x8060b9c;
              FUN_08050ebc();
              local_70 = (undefined1  [4])0x0;
              local_6c = (BigNum *)0x0;
              pBStack_130 = (BigNum *)0x8060bb2;
              local_88 = pBVar25;
              FUN_08050ebc();
              goto LAB_08060bbb;
            }
          }
          auVar37 = (undefined1  [4])0x0;
LAB_08060bbb:
          pBStack_130 = (BigNum *)0x8060bcc;
          local_8c = (BigNum *)auVar37;
          cVar7 = FUN_08050c9a();
          ppBVar32 = &local_12c;
          if (((cVar7 == '\0') &&
              (ppBVar32 = &local_12c, *(int *)(*(int *)(param_1 + 0x30) + 0x2c) != 0)) &&
             (piVar14 = *(int **)(*(int *)(param_1 + 0x30) + 0x28), ppBVar32 = &local_12c,
             piVar14 != (int *)0x0)) {
            pBStack_130 = local_104 + 0x9c;
            (**(code **)(*piVar14 + 8))();
            ppBVar32 = &pBStack_130;
            FUN_08050ebc();
            local_88 = local_6c;
            auVar37 = local_70;
            local_6c = (BigNum *)0x0;
            local_70 = (undefined1  [4])0x0;
            FUN_08050ebc();
          }
          *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060c3f;
          local_8c = (BigNum *)auVar37;
          pBVar15 = (BigNum *)FUN_08050c9a();
          if ((char)pBVar15 == '\0') {
            *(BigNum **)((int)ppBVar32 + 0xfffffffc) = pBVar25;
            *(BigNum **)((int)ppBVar32 + 0xfffffff8) = pBVar25;
            *(char **)((int)ppBVar32 + 0xfffffff4) = "could not locate OCSP issuer certificate";
            *(undefined1 **)((int)ppBVar32 + 0xfffffff0) = local_70;
            *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060c58;
            string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                           *(char **)((int)ppBVar32 + 0xfffffff4));
            *(undefined1 **)((int)ppBVar32 + 0xfffffff0) = local_70;
            *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060c70;
            FUN_0805e57c();
            *(undefined1 **)((int)ppBVar32 + 0xfffffff0) = local_70;
            *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060c78;
            string::freeptr();
          }
          else {
            local_70 = (undefined1  [4])0x0;
            local_6c = (BigNum *)0x0;
            local_68 = (BigNum *)0x0;
            local_64 = (BigNum *)0x0;
            local_60 = 0;
            local_5c = 0;
            local_58 = 0;
            local_54 = 0;
            uStack_52 = 0;
            local_50 = 0;
            uStack_4e = 0;
            local_4c[0] = '\0';
            local_4c[1] = '\0';
            local_4c[2] = '\0';
            local_4c[3] = '\0';
            local_4c[4] = '\0';
            local_4c[5] = '\0';
            local_4c[6] = '\0';
            local_4c[7] = '\0';
            local_44 = 0;
            local_40 = (undefined4 *)0x0;
            local_3c = (undefined4 *)0x0;
            local_38 = 0;
            local_34 = 0;
            local_30 = 0;
            local_2c = 0;
            local_28 = 0;
            local_24 = 0;
            local_20 = 0;
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = extraout_ECX_01;
            *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = local_78[0];
            *(int *)((int)ppBVar32 + 0xfffffff4) = local_7c;
            *(undefined1 **)((int)ppBVar32 + 0xfffffff0) = local_70;
            *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060cd3;
            OcspResponse::parserFromBlob();
            pBVar15 = local_68;
            if ((char)local_68 == '\0') {
              *(BigNum **)((int)ppBVar32 + 0xfffffffc) = pBVar25;
              *(BigNum **)((int)ppBVar32 + 0xfffffff8) = pBVar25;
              *(char **)((int)ppBVar32 + 0xfffffff4) = "malformed OCSP response";
              *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
              *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060cf2;
              string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                             *(char **)((int)ppBVar32 + 0xfffffff4));
              *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
              *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d0a;
              FUN_0805e57c();
              *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
            }
            else {
              *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_a4;
              *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d21;
              string::string(*(string **)((int)ppBVar32 + 0xfffffff0));
              *(BigNum ***)((int)ppBVar32 + 0xffffffec) = &local_a4;
              *(undefined1 (*) [4])((int)ppBVar32 + 0xffffffe8) = auVar37;
              *(BigNum **)((int)ppBVar32 + 0xffffffe4) = local_104;
              *(undefined1 **)((int)ppBVar32 + 0xffffffe0) = local_70;
              *(undefined4 *)((int)ppBVar32 + 0xffffffdc) = 0x8060d32;
              iVar21 = OcspResponse::verify
                                 (*(X509Certificate **)((int)ppBVar32 + 0xffffffe0),
                                  *(X509Certificate **)((int)ppBVar32 + 0xffffffe4),
                                  *(string **)((int)ppBVar32 + 0xffffffe8));
              if (iVar21 != 1) {
                iVar39 = *(int *)local_a4;
                if (iVar21 == 2) {
                  if (iVar39 == 0) {
                    *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = extraout_ECX_02;
                    *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = extraout_ECX_02;
                    *(char **)((int)ppBVar32 + 0xfffffff4) = "certificate revoked by OCSP";
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d64;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(char **)((int)ppBVar32 + 0xfffffff4));
                  }
                  else {
                    *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 2;
                    *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 2;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff4) = &local_a4;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d6f;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(string **)((int)ppBVar32 + 0xfffffff4));
                  }
                  *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                }
                else if (iVar21 == 3) {
                  if (iVar39 == 0) {
                    *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0;
                    *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 0;
                    *(char **)((int)ppBVar32 + 0xfffffff4) = "certificate status unknown per OCSP";
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060db0;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(char **)((int)ppBVar32 + 0xfffffff4));
                  }
                  else {
                    *(int *)((int)ppBVar32 + 0xfffffffc) = iVar39;
                    *(int *)((int)ppBVar32 + 0xfffffff8) = iVar39;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff4) = &local_a4;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060dbb;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(string **)((int)ppBVar32 + 0xfffffff4));
                  }
                  *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                }
                else {
                  if (iVar39 == 0) {
                    *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0;
                    *(undefined4 *)((int)ppBVar32 + 0xfffffff8) = 0;
                    *(char **)((int)ppBVar32 + 0xfffffff4) = "OCSP response could not be verified";
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060de0;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(char **)((int)ppBVar32 + 0xfffffff4));
                  }
                  else {
                    *(int *)((int)ppBVar32 + 0xfffffffc) = iVar39;
                    *(int *)((int)ppBVar32 + 0xfffffff8) = iVar39;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff4) = &local_a4;
                    *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                    *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060deb;
                    string::string(*(string **)((int)ppBVar32 + 0xfffffff0),
                                   *(string **)((int)ppBVar32 + 0xfffffff4));
                  }
                  *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                }
                *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d8b;
                FUN_0805e57c();
                *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_9c;
                pBVar15 = (BigNum *)0x0;
                *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060d95;
                string::freeptr();
              }
              *(BigNum ***)((int)ppBVar32 + 0xfffffff0) = &local_a4;
            }
            *(undefined4 *)((int)ppBVar32 + 0xffffffec) = 0x8060e10;
            string::freeptr();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e1b;
            FUN_08050ebc();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e23;
            FUN_08050f2a();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e2b;
            FUN_0805ebd2();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e33;
            FUN_0805091c();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e3b;
            FUN_0805091c();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e43;
            FUN_0805091c();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e4b;
            FUN_0805091c();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e53;
            FUN_0805091c();
            *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e5b;
            FUN_0805091c();
          }
          *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e66;
          FUN_08050ebc();
        }
        *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e71;
        FUN_08050ebc();
        if ((char)pBVar15 != '\0') {
          *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e7f;
          FUN_0805091c();
          return pBVar15;
        }
      }
    }
    *(undefined4 *)((int)ppBVar32 + 0xfffffffc) = 0x8060e8c;
    FUN_0805091c();
  }
  else {
switchD_080600bd_caseD_5:
    local_70 = (undefined1  [4])0x10;
    piVar26 = (int *)&stack0xfffffec4;
    string::string((string *)&local_6c);
LAB_08060171:
    piVar26[-1] = 0x8060179;
    FUN_0805a4bc();
    *piVar26 = (int)&local_6c;
LAB_0806017c:
    *(undefined4 *)((int)piVar26 + -4) = 0x8060181;
    string::freeptr();
  }
  return (BigNum *)0x0;
code_r0x0806054d:
  local_104 = (BigNum *)(*(int *)(param_1 + 0x15c) + 0x44);
  (**(code **)(*(int *)*puVar17 + 0xc))();
  iVar21 = FUN_08059468();
  puVar17 = puVar17 + 1;
  if (iVar21 == 0) goto code_r0x08060583;
  goto LAB_08060549;
code_r0x08060583:
  local_80 = (BigNum *)0x13;
  string::string((string *)&local_7c);
  goto LAB_080628ba;
}


