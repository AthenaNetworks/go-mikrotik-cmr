
undefined4 FUN_0808dcb4(int *param_1,uint param_2,int param_3,int *param_4)

{
  int *piVar1;
  undefined1 uVar2;
  char cVar3;
  Timer *pTVar4;
  undefined4 *puVar5;
  ostream *poVar6;
  undefined4 uVar7;
  int iVar8;
  undefined4 extraout_ECX;
  undefined4 extraout_ECX_00;
  undefined4 extraout_ECX_01;
  int **ppiVar9;
  int *piVar10;
  undefined1 *puVar11;
  undefined1 *puVar12;
  undefined1 *puVar13;
  uint *puVar15;
  code *pcVar16;
  byte bVar17;
  uint auStackY_cc [3];
  undefined1 *puVar18;
  char *pcVar19;
  undefined1 auStack_ac [8];
  int local_a4;
  int local_a0;
  undefined4 local_90;
  int local_8c [4];
  undefined4 local_7c;
  undefined4 local_78;
  int *local_74;
  undefined4 *local_70;
  undefined4 local_6c;
  undefined4 local_68;
  char local_64;
  int local_60;
  int local_5c;
  int local_58;
  undefined4 local_54;
  undefined4 local_50;
  int local_4c;
  int local_48;
  undefined4 local_44;
  int *local_40;
  undefined4 *local_3c;
  int local_38;
  undefined4 local_34;
  undefined1 *puVar14;
  
  bVar17 = 0;
  if (param_2 == 0x44) {
    local_5c = *(int *)(param_3 + 4);
    local_58 = *(int *)(param_3 + 8);
    local_60 = 0;
    local_54 = *(undefined4 *)(param_3 + 0xc);
    if (param_1[0x32] == 3) {
      poVar6 = (ostream *)operator<<();
      endl(poVar6);
      local_40 = (int *)0x0;
      nv::Timer::operator=((Timer *)(param_1 + 0x14),(Timer *)&local_40);
      nv::Timer::~Timer((Timer *)&local_40);
      local_6c = 6;
      param_1[0x15] = 0;
      FUN_0805e79c();
      local_40 = (int *)0x0;
      local_3c = (undefined4 *)local_5c;
      local_38 = local_58;
      local_34 = local_54;
      cVar3 = FUN_0808669c();
      FUN_08050ec6();
      FUN_0805e336();
      puVar11 = auStack_ac;
      if (cVar3 == '\0') {
        local_8c[2] = 0;
        local_8c[3] = 0;
        FUN_08051b82();
        cVar3 = FUN_08086816();
        if (cVar3 == '\0') {
          pcVar16 = *(code **)(*param_1 + 0x10);
          string::string((string *)&local_40,"auth failed (3)");
          (*pcVar16)();
          string::freeptr();
          puVar13 = auStack_ac;
        }
        else {
          local_50 = 0;
          local_48 = local_8c[3] - local_8c[2];
          local_4c = local_8c[2];
          local_44 = 0;
          FUN_0805e77a();
          local_40 = (int *)0x0;
          local_38 = local_48;
          local_3c = (undefined4 *)local_4c;
          local_34 = local_44;
          cVar3 = FUN_0808669c();
          FUN_08050ec6();
          FUN_0805e336();
          if (cVar3 == '\0') {
            FUN_080882fc();
            FUN_08086d64();
            local_90 = 0;
            FUN_08086d8a();
            FUN_08051b82();
            cVar3 = FUN_08086eda();
            if (cVar3 == '\0') {
LAB_0808e52d:
              pcVar16 = *(code **)(*param_1 + 0x10);
              goto LAB_0808e537;
            }
            FUN_08051b82();
            cVar3 = FUN_08089ce6();
            if (cVar3 == '\0') goto LAB_0808e52d;
            FUN_08051b82();
            poVar6 = (ostream *)operator<<();
            operator<<(poVar6,": peer id: ");
            FUN_08086420();
            poVar6 = (ostream *)FUN_08051836();
            endl(poVar6);
            string::freeptr();
            puVar18 = (undefined1 *)param_1[0x31];
            if ((undefined4 *)puVar18 == (undefined4 *)0x0) {
              pcVar16 = *(code **)(*param_1 + 0x10);
              pcVar19 = "no peer";
              goto LAB_0808e5de;
            }
            FUN_08087ee2();
            pTVar4 = (Timer *)&local_40;
            for (iVar8 = 8; iVar8 != 0; iVar8 = iVar8 + -1) {
              *(undefined4 *)puVar18 = *(undefined4 *)pTVar4;
              pTVar4 = pTVar4 + (uint)bVar17 * -8 + 4;
              puVar18 = (undefined1 *)((int)puVar18 + ((uint)bVar17 * -2 + 1) * 4);
            }
            iVar8 = FUN_08051968();
            uVar7 = extraout_ECX;
            if (iVar8 == 2) {
              puVar18 = cout;
              poVar6 = (ostream *)operator<<();
              poVar6 = (ostream *)operator<<(poVar6,": received addrs");
              endl(poVar6);
              m3::message::read::get((uint)&local_40);
              local_74 = param_1;
              m3::each(local_40);
              puVar11 = &stack0xffffff50;
              uVar7 = extraout_ECX_00;
            }
            if (DAT_0809a89c == 1) {
              if (((byte)runtimeDFLAGS & 2) != 0) goto LAB_0808e68f;
            }
            else if (-1 < DAT_0809a89c + -1) {
LAB_0808e68f:
              if (param_1[0x32] != 3) {
                *(undefined1 **)(puVar11 + -4) = puVar18;
                *(undefined1 **)(puVar11 + -8) = puVar18;
                *(undefined **)(puVar11 + -0xc) = &DAT_080900e8;
                *(undefined1 **)(puVar11 + -0x10) = cerr;
                *(undefined4 *)(puVar11 + -0x14) = 0x808e6ac;
                uVar7 = operator<<(*(ostream **)(puVar11 + -0x10),*(char **)(puVar11 + -0xc));
                *(undefined **)(puVar11 + -0xc) = &DAT_0809551a;
                *(undefined4 *)(puVar11 + -0x10) = uVar7;
                *(undefined4 *)(puVar11 + -0x14) = 0x808e6b9;
                uVar7 = operator<<(*(ostream **)(puVar11 + -0x10),*(char **)(puVar11 + -0xc));
                *(undefined4 *)(puVar11 + -0x10) = uVar7;
                *(undefined4 *)(puVar11 + -0x14) = 0x808e6c1;
                endl(*(ostream **)(puVar11 + -0x10));
                *(undefined **)(puVar11 + -0x10) = &DAT_080900e5;
                *(undefined4 *)(puVar11 + -0x14) = 0x808e6cd;
                FUN_080865c4();
                uVar7 = extraout_ECX_01;
              }
            }
            *(undefined4 *)(puVar11 + -4) = uVar7;
            *(undefined4 *)(puVar11 + -8) = 0x16;
            *(char **)(puVar11 + -0xc) = "PKEXClntImpl::sendMsg4(const PKEX::Context&)";
            *(undefined1 **)(puVar11 + -0x10) = cout;
            *(undefined4 *)(puVar11 + -0x14) = 0x808e6eb;
            uVar7 = operator<<();
            *(undefined4 *)(puVar11 + -0x10) = uVar7;
            *(undefined4 *)(puVar11 + -0x14) = 0x808e6f3;
            endl(*(ostream **)(puVar11 + -0x10));
            local_7c = 0;
            local_78 = 0;
            *(undefined4 *)(puVar11 + -0x14) = 0x808e707;
            FUN_08089a26();
            local_8c[0] = 0;
            *(undefined4 *)(puVar11 + -0x14) = 0;
            *(undefined4 *)(puVar11 + -0x18) = 0x20;
            *(int **)(puVar11 + -0x1c) = param_1 + 0x18;
            *(int ***)(puVar11 + -0x20) = &local_40;
            *(undefined4 *)(puVar11 + -0x24) = 0x808e724;
            ECPoint::toBin(*(uint *)(puVar11 + -0x20),(bool)puVar11[-0x1c]);
            puVar12 = puVar11 + -4;
            *(undefined4 *)(puVar11 + -8) = 0x808e735;
            FUN_0805d9b6();
            *(undefined4 *)(puVar11 + -8) = 0x808e73c;
            FUN_0805da10();
            *(undefined4 *)(puVar11 + -8) = 0x808e74b;
            FUN_0805d9b6();
            *(undefined4 *)(puVar11 + -8) = 0x808e758;
            FUN_0805d6b8();
            *(undefined4 *)(puVar11 + -8) = 0x808e766;
            FUN_0805d9b6();
            *(undefined4 *)(puVar11 + -8) = 0x808e76d;
            FUN_0805da10();
            if (param_1[0x12] != param_1[0x13]) {
              *(undefined4 *)(puVar11 + -8) = 1;
              *(undefined4 *)(puVar11 + -0xc) = 9;
              *(int **)(puVar11 + -0x10) = local_8c;
              *(int ***)(puVar11 + -0x14) = &local_40;
              *(undefined4 *)(puVar11 + -0x18) = 0x808e783;
              m3::message::putvc(*(uint *)(puVar11 + -0x14),*(uint *)(puVar11 + -0x10));
              puVar12 = puVar11 + -8;
              piVar1 = (int *)param_1[0x13];
              for (piVar10 = (int *)param_1[0x12]; piVar10 != piVar1; piVar10 = piVar10 + 2) {
                local_a4 = *piVar10;
                local_a0 = piVar10[1] - local_a4;
                *(undefined4 *)(puVar11 + -0xc) = 1;
                *(int *)(puVar11 + -0x10) = local_a0;
                *(int *)(puVar11 + -0x14) = local_a4;
                *(int ***)(puVar11 + -0x18) = &local_40;
                *(undefined4 *)(puVar11 + -0x1c) = 0x808e7c0;
                m3::message::WriteHandle::putChunk();
              }
              *(undefined4 *)(puVar11 + -0xc) = 0x808e7cd;
              FUN_08051114();
            }
            *(undefined4 *)(puVar12 + -4) = 0x808e7ea;
            FUN_0808636a();
            *(undefined4 *)(puVar12 + -4) = 0x808e7f7;
            FUN_080510e8();
            *(undefined4 *)(puVar12 + -4) = 0x808e804;
            cVar3 = FUN_08086728();
            *(undefined4 *)(puVar12 + -4) = 0x808e80d;
            FUN_08050ec6();
            if (cVar3 == '\0') {
              iVar8 = *param_1;
              pcVar16 = *(code **)(iVar8 + 0x10);
              *(int *)(puVar12 + -4) = iVar8;
              *(int *)(puVar12 + -8) = iVar8;
              *(char **)(puVar12 + -0xc) = "auth failed (4)";
              *(int ***)(puVar12 + -0x10) = &local_40;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e828;
              string::string(*(string **)(puVar12 + -0x10),*(char **)(puVar12 + -0xc));
              *(int ***)(puVar12 + -0xc) = &local_40;
              *(int **)(puVar12 + -0x10) = param_1;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e830;
              (*pcVar16)();
              *(int ***)(puVar12 + -0x10) = &local_40;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e838;
              string::freeptr();
            }
            else {
              local_8c[1] = 0;
              *(undefined4 *)(puVar12 + -4) = 0x808e855;
              FUN_0805d9b6();
              *(int **)(puVar12 + -0x10) = param_4;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e86e;
              FUN_08089616();
              pcVar16 = *(code **)(*param_1 + 0x10);
              *(undefined **)(puVar12 + -0xc) = &DAT_08092eed;
              *(int ***)(puVar12 + -0x10) = &local_40;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e883;
              string::string(*(string **)(puVar12 + -0x10),*(char **)(puVar12 + -0xc));
              *(int ***)(puVar12 + -0xc) = &local_40;
              *(int **)(puVar12 + -0x10) = param_1;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e88b;
              (*pcVar16)();
              *(int ***)(puVar12 + -0x10) = &local_40;
              *(undefined4 *)(puVar12 + -0x14) = 0x808e893;
              string::freeptr();
              *(undefined4 *)(puVar12 + -0x14) = 0x808e89e;
              FUN_08051114();
            }
            puVar14 = puVar12 + -0x10;
            *(undefined4 *)(puVar12 + -0x14) = 0x808e8a6;
            FUN_0805da10();
            *(undefined4 *)(puVar12 + -0x14) = 0x808e8b1;
            FUN_08051114();
            *(undefined4 *)(puVar12 + -0x14) = 0x808e8b9;
            FUN_0805da10();
          }
          else {
            poVar6 = (ostream *)operator<<();
            poVar6 = (ostream *)operator<<(poVar6,": missing fields in decrypted msg");
            endl(poVar6);
            pcVar16 = *(code **)(*param_1 + 0x10);
LAB_0808e537:
            pcVar19 = "auth failed (3)";
LAB_0808e5de:
            string::string((string *)&local_40,pcVar19);
            puVar14 = &stack0xffffff44;
            (*pcVar16)();
            string::freeptr();
          }
          puVar13 = puVar14 + 0x10;
          *(undefined4 *)(puVar14 + 0xc) = 0x808e8c4;
          FUN_08050ec6();
        }
        ppiVar9 = &local_40;
        *(undefined4 *)(puVar13 + -4) = 0x808e8cc;
        FUN_0805da10();
LAB_0808e8cc:
        uVar7 = CONCAT31((int3)((uint)ppiVar9 >> 8),1);
        goto LAB_0808e8d1;
      }
      poVar6 = (ostream *)operator<<();
      pcVar19 = ": invalid msg3 received";
LAB_0808e3b9:
      poVar6 = (ostream *)operator<<(poVar6,pcVar19);
      endl(poVar6);
    }
LAB_0808e2f7:
    puVar13 = auStack_ac;
    uVar7 = 0;
  }
  else {
    if (0x44 < param_2) {
LAB_0808e8dc:
      uVar7 = operator<<();
      puVar15 = (uint *)&stack0xffffff48;
      goto LAB_0808e91a;
    }
    if (param_2 != 0x41) {
      if (param_2 != 0x42) goto LAB_0808e8dc;
      local_58 = *(int *)(param_3 + 8);
      local_60 = 0;
      local_5c = *(int *)(param_3 + 4);
      local_54 = *(undefined4 *)(param_3 + 0xc);
      if (param_1[0x32] != 2) goto LAB_0808e2f7;
      poVar6 = (ostream *)operator<<();
      endl(poVar6);
      local_40 = (int *)0x0;
      nv::Timer::operator=((Timer *)(param_1 + 0x35),(Timer *)&local_40);
      nv::Timer::~Timer((Timer *)&local_40);
      local_40 = (int *)0x0;
      nv::Timer::operator=((Timer *)(param_1 + 0x33),(Timer *)&local_40);
      nv::Timer::~Timer((Timer *)&local_40);
      local_6c = 0;
      local_68 = 4;
      FUN_0805e79c();
      local_40 = (int *)0x0;
      local_38 = local_58;
      local_3c = (undefined4 *)local_5c;
      local_34 = local_54;
      cVar3 = FUN_0808669c();
      FUN_08050ec6();
      FUN_0805e336();
      if (cVar3 != '\0') {
        poVar6 = (ostream *)operator<<();
        pcVar19 = ": invalid msg1 received";
        goto LAB_0808e3b9;
      }
      FUN_080519c4();
      if (local_40 != (int *)0x1) {
        poVar6 = (ostream *)operator<<();
        poVar6 = (ostream *)operator<<(poVar6,": protocol version missmatch");
        endl(poVar6);
        ppiVar9 = *(int ***)(*param_1 + 0x10);
LAB_0808e1cf:
        string::string((string *)&local_40,"auth failed (1)");
        (*(code *)ppiVar9)();
        string::freeptr();
        puVar13 = auStack_ac;
        goto LAB_0808e8cc;
      }
      FUN_080882fc();
      FUN_08086d64();
      local_6c = 0;
      FUN_08086d8a();
      FUN_08051b82();
      cVar3 = FUN_08086eda();
      local_a4 = CONCAT31(local_a4._1_3_,cVar3);
      if (cVar3 == '\0') {
        ppiVar9 = *(int ***)(*param_1 + 0x10);
        goto LAB_0808e1cf;
      }
      FUN_0808a872();
      WCurve::fromEC((ECPoint *)&local_40);
      puVar13 = &stack0xffffff50;
      uVar7 = FUN_0808bfd4();
      FUN_08086dc4();
      FUN_08086d42();
      if ((char)uVar7 == '\0') {
        pcVar16 = *(code **)(*param_1 + 0x10);
        string::string((string *)&local_40,"auth failed (1)");
        (*pcVar16)(param_1);
        string::freeptr();
        uVar7 = CONCAT31((int3)((uint)pcVar16 >> 8),(char)local_a4);
        puVar13 = &stack0xffffff50;
      }
      else {
        FUN_0808dabc();
      }
      goto LAB_0808e8d1;
    }
    FUN_08051b82();
    if (local_64 == '\0') {
      return 0;
    }
    uVar7 = FUN_0808c8d2();
    if ((char)uVar7 == '\0') {
      pcVar16 = *(code **)(*param_1 + 0x10);
      string::string((string *)&local_40,"session overlap detected");
      (*pcVar16)();
      string::freeptr();
      return uVar7;
    }
    local_48 = *(int *)(param_3 + 8);
    local_50 = 0;
    local_4c = *(int *)(param_3 + 4);
    local_44 = *(undefined4 *)(param_3 + 0xc);
    if ((uint)param_1[0x32] < 3) {
      FUN_0805e77a();
      local_40 = (int *)0x0;
      local_3c = (undefined4 *)local_4c;
      local_38 = local_48;
      local_34 = local_44;
      uVar2 = FUN_0808669c();
      local_a4 = CONCAT31(local_a4._1_3_,uVar2);
      FUN_08050ec6();
      FUN_0805e336();
      if ((char)local_a4 != '\0') {
        poVar6 = (ostream *)operator<<();
        poVar6 = (ostream *)operator<<(poVar6,": invalid discovery response received");
        endl(poVar6);
        goto LAB_0808dd69;
      }
      local_60 = param_1[0x29];
      local_a4 = param_1[0x2a];
      local_5c = local_a4 - local_60;
      FUN_08051b82();
      cVar3 = FUN_080854f2();
      local_a4 = CONCAT31(local_a4._1_3_,cVar3);
      if (cVar3 == '\0') {
        pcVar16 = *(code **)(*param_1 + 0x10);
        string::string((string *)&local_40,"discovery failed: invalid hash received");
        (*pcVar16)();
        string::freeptr();
      }
      else {
        FUN_08051b82();
        poVar6 = (ostream *)operator<<();
        operator<<(poVar6,": nonce: ");
        FUN_08086420();
        poVar6 = (ostream *)FUN_08051836();
        endl(poVar6);
        string::freeptr();
        FUN_0805d2b2();
        FUN_0808693e();
        pTVar4 = malloc(8);
        BigNum::BigNum();
        local_7c = 0;
        FUN_08086f54();
        FUN_08086f54();
        FUN_0805da10();
        FUN_0805da10();
        if (param_1[0x32] == 0) {
          poVar6 = (ostream *)operator<<();
          poVar6 = (ostream *)operator<<(poVar6,": server found");
          endl(poVar6);
          local_40 = (int *)0x0;
          nv::Timer::operator=((Timer *)(param_1 + 0x35),(Timer *)&local_40);
          nv::Timer::~Timer((Timer *)&local_40);
          param_1[0x32] = 1;
          if (param_1[6] != 0) {
            local_40 = (int *)0x0;
            local_3c = malloc(8);
            *local_3c = &PTR_LAB_08098924;
            local_3c[1] = param_1;
            nv::Timer::Timer();
            nv::Timer::operator=((Timer *)(param_1 + 0x33),(Timer *)&local_74);
            nv::Timer::~Timer((Timer *)&local_74);
            FUN_08051296();
          }
          local_40 = param_1;
          local_3c = (undefined4 *)*param_4;
          local_38 = param_4[1];
          FUN_0808554a();
          local_74 = (int *)0x0;
          puVar5 = malloc(0x10);
          *puVar5 = &PTR_LAB_0809893c;
          FUN_08085556();
          pTVar4 = (Timer *)&local_7c;
          local_70 = puVar5;
          nv::Timer::Timer();
          nv::Timer::operator=((Timer *)(param_1 + 0x34),pTVar4);
          nv::Timer::~Timer(pTVar4);
          FUN_08051296();
          FUN_08088998();
        }
        uVar7 = CONCAT31((int3)((uint)pTVar4 >> 8),(char)local_a4);
      }
    }
    else {
LAB_0808dd69:
      uVar7 = 0;
    }
    puVar13 = auStack_ac;
  }
LAB_0808e8d1:
  *(undefined4 *)(puVar13 + -4) = 0x808e8d6;
  FUN_08050ec6();
  if ((char)uVar7 != '\0') {
    return uVar7;
  }
  *(undefined4 *)(puVar13 + -4) = uVar7;
  *(undefined4 *)(puVar13 + -8) = 0x1e;
  *(char **)(puVar13 + -0xc) =
       "bool PKEXClntImpl::handleFrame(unsigned int, m3::msg_view, const PKEX::Context&)";
  *(undefined1 **)(puVar13 + -0x10) = cout;
  *(undefined4 *)(puVar13 + -0x14) = 0x808e913;
  uVar7 = operator<<();
  puVar15 = (uint *)(puVar13 + -0xc);
  *(char **)(puVar13 + -0xc) = ": unhandled msg ";
LAB_0808e91a:
  puVar15[-1] = uVar7;
  puVar15[-2] = 0x808e922;
  uVar7 = operator<<((ostream *)puVar15[-1],(char *)*puVar15);
  *puVar15 = param_2;
  puVar15[-1] = uVar7;
  puVar15[-2] = 0x808e92d;
  uVar7 = operator<<((ostream *)puVar15[-1],*puVar15);
  puVar15[-1] = uVar7;
  puVar15[-2] = 0x808e935;
  endl((ostream *)puVar15[-1]);
  return 0;
}


