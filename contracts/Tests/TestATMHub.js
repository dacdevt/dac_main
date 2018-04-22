const dacHub= artifacts.require("dacHub");
const dacToken = artifacts.require("dacToken");

contract('发布广告/扫码/删除广告 : ', function(accounts) {
  let dac;
  let dachub;
  var dac = 100000000;
  var paltform  = accounts[0];
  var advertiser = accounts[1];
  var screen = accounts[2];
  var user = accounts[3];

  it("Case1: 1个广告,限制5次,用户扫码6次,删除广告", async function() {

    dac = await dacToken.new(100000*dac);
    dachub = await dacHub.new(dac.address);

    console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
  
    await dac.transfer(advertiser,100000*dac);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

    await dac.approve(dachub.address,9*dac,{from: advertiser});
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 9*dac);

    /*发布广告: id=12121 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
    var scanCount = 5;
    var mediaID1 = 12121;
    await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,0]);
    console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));
    /* mediaStrategy的数据结构
        uint mediaID; 0
        uint mediaIndex; 1
        address advertiser; 2
        uint totaldacRewards; 3
        uint currentTotal; 4
        uint totalCount; 5
        uint currentCount; 6
        uint userPayPrice;    7
        uint ratio1;    //rewards ratio 8
        uint ratio2; 
        uint ratio3; 
        bool enableIssueFlag; 9
    */
    assert.equal((await dachub.mediaStrategy(mediaID1))[0],mediaID1);
    assert.equal((await dachub.mediaStrategy(mediaID1))[2],advertiser);
    assert.equal((await dachub.mediaStrategy(mediaID1))[3].toNumber(),1*dac*1.8*scanCount);
    assert.equal((await dachub.mediaStrategy(mediaID1))[6],5);
    assert.equal((await dachub.mediaStrategy(mediaID1))[8],50);
    assert.equal((await dachub.mediaStrategy(mediaID1))[9],30);
    assert.equal((await dachub.mediaStrategy(mediaID1))[10],0);
    assert.equal((await dachub.mediaStrategy(mediaID1))[11],true);
    /* 1次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 1*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.5*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 0.3*dac);
    /* 2次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 2*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 1*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 0.6*dac);
    /* 3次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 3*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 1.5*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 0.9*dac);
    /* 4次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 4*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 2.0*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 1.2*dac);
    /* 5次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 5*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 2.5*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 1.5*dac);
    /* 6次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 5*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 2.5*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 1.5*dac);
    assert.equal((await dachub.mediaStrategy(mediaID1))[6],0);
    assert.equal((await dachub.mediaStrategy(mediaID1))[4].toNumber(),0*dac);
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0*dac);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99991*dac);

    /*删除广告 */
    await dachub.deleteAdvertise(mediaID1);
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99991*dac);
  });

  it("Case2: 1个广告,限制5次,用户扫码2次,删除广告", async function() {
    
    dac = await dacToken.new(100000*dac);
    dachub = await dacHub.new(dac.address);

    console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
    
    await dac.transfer(advertiser,100000*dac);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

    await dac.approve(dachub.address,10*dac,{from: advertiser});
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 10*dac);

    /*发布广告: id=12121 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
    var scanCount = 5;
    var mediaID1 = 12121;
    await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
    console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));
    /* mediaStrategy的数据结构
        uint mediaID; 0
        uint mediaIndex; 1
        address advertiser; 2
        uint totaldacRewards; 3
        uint currentTotal; 4
        uint totalCount; 5
        uint currentCount; 6
        uint userPayPrice;    7
        uint ratio1;    //rewards ratio 8
        uint ratio2; 
        uint ratio3; 
        bool enableIssueFlag; 9
    */
    assert.equal((await dachub.mediaStrategy(mediaID1))[0],mediaID1);
    assert.equal((await dachub.mediaStrategy(mediaID1))[2],advertiser);
    assert.equal((await dachub.mediaStrategy(mediaID1))[3].toNumber(),1*dac*2*scanCount);
    assert.equal((await dachub.mediaStrategy(mediaID1))[6],5);
    /* 1次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 1*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.5*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 0.3*dac);
    /* 2次扫码 */
    await dachub.issueReward(user,screen,mediaID1,100);
    assert.equal((await dac.balanceOf(user)).toNumber(), 2*dac);
    assert.equal((await dac.balanceOf(paltform)).toNumber(), 1*dac);
    assert.equal((await dac.balanceOf(screen)).toNumber(), 0.6*dac);
    assert.equal((await dachub.mediaStrategy(mediaID1))[6],3);
    assert.equal((await dachub.mediaStrategy(mediaID1))[4].toNumber(),6.4*dac);
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 6.4*dac);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99996.4*dac);

    /*删除广告 */
    await dachub.deleteAdvertise(mediaID1);
    assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0);
    assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99996.4*dac);
    });

    it("Case3: 1个广告,限制5次,删除广告后用户扫码1次,", async function() {
    
        dac = await dacToken.new(100000*dac);
        dachub = await dacHub.new(dac.address);

        console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
        
        await dac.transfer(advertiser,100000*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        await dac.approve(dachub.address,10*dac,{from: advertiser});
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 10*dac);

        /*发布广告: id=12121 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
        var scanCount = 5;
        var mediaID1 = 12121;
        await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));
        /* mediaStrategy的数据结构
        uint mediaID; 0
        uint mediaIndex; 1
        address advertiser; 2
        uint totaldacRewards; 3
        uint currentTotal; 4
        uint totalCount; 5
        uint currentCount; 6
        uint userPayPrice;    7
        uint ratio1;    //rewards ratio 8
        uint ratio2; 
        uint ratio3; 
        bool enableIssueFlag; 9
        */
        assert.equal((await dachub.mediaStrategy(mediaID1))[0],mediaID1);
        assert.equal((await dachub.mediaStrategy(mediaID1))[2],advertiser);
        assert.equal((await dachub.mediaStrategy(mediaID1))[3].toNumber(),1*dac*2*scanCount);
        assert.equal((await dachub.mediaStrategy(mediaID1))[6],5);

        /*删除广告 */
        await dachub.deleteAdvertise(mediaID1);
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0);
        
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,100);
        assert.equal((await dac.balanceOf(user)).toNumber(), 0);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0);
        assert.equal((await dachub.mediaStrategy(mediaID1))[6],0);
        assert.equal((await dachub.mediaStrategy(mediaID1))[4].toNumber(),0);
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);
    });

    it("Case4: 1个广告,限制5次,用户注意力度量50%扫码1次,用户注意力度量10%扫码1次", async function() {
        
        dac = await dacToken.new(100000*dac);
        dachub = await dacHub.new(dac.address);

        console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
        
        await dac.transfer(advertiser,100000*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        await dac.approve(dachub.address,10*dac,{from: advertiser});
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 10*dac);

        /*发布广告: id=12121 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
        var scanCount = 5;
        var mediaID1 = 12121;
        await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));
        /* mediaStrategy的数据结构
        uint mediaID; 0
        uint mediaIndex; 1
        address advertiser; 2
        uint totaldacRewards; 3
        uint currentTotal; 4
        uint totalCount; 5
        uint currentCount; 6
        uint userPayPrice;    7
        uint ratio1;    //rewards ratio 8
        uint ratio2; 
        uint ratio3; 
        bool enableIssueFlag; 9
        */
        assert.equal((await dachub.mediaStrategy(mediaID1))[0],mediaID1);
        assert.equal((await dachub.mediaStrategy(mediaID1))[2],advertiser);
        assert.equal((await dachub.mediaStrategy(mediaID1))[3].toNumber(),1*dac*2*scanCount);
        assert.equal((await dachub.mediaStrategy(mediaID1))[6],5);

        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,50);
        assert.equal((await dac.balanceOf(user)).toNumber(), 0.5*dac);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.25*dac);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0.15*dac);
        /* 2次扫码 */
        await dachub.issueReward(user,screen,mediaID1,10);
        assert.equal((await dac.balanceOf(user)).toNumber(), 0.6*dac);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.3*dac);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0.18*dac);
        
        assert.equal((await dachub.mediaStrategy(mediaID1))[6],3);
        assert.equal((await dachub.mediaStrategy(mediaID1))[4].toNumber(),8.92*dac);
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 8.92*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99998.92*dac);

        /*删除广告 */
        await dachub.deleteAdvertise(mediaID1);
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 0);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 99998.92*dac);
    });
    it("Case5: 3个广告,限制5次,用户分别扫码1次", async function() {
        
        dac = await dacToken.new(100000*dac);
        dachub = await dacHub.new(dac.address);

        console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
        
        await dac.transfer(advertiser,100000*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        await dac.approve(dachub.address,30*dac,{from: advertiser});
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 30*dac);

        /*发布广告: id=12121.12122,12123 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
        var scanCount = 5;
        var mediaID1 = 12121;
        var mediaID2 = 12122;
        var mediaID3 = 12123;
        await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));
        await dachub.publishAdvertise(advertiser,mediaID2,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID2)));
        await dachub.publishAdvertise(advertiser,mediaID3,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID3)));

        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,50);
        await dachub.issueReward(user,screen,mediaID2,50);
        await dachub.issueReward(user,screen,mediaID3,50);
        assert.equal((await dac.balanceOf(user)).toNumber(), 1.5*dac);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.75*dac);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0.45*dac);

    });

    it("Case6: 1个广告,限制5次,用户扫码1次,追加扫码5次", async function() {
        
        dac = await dacToken.new(100000*dac);
        dachub = await dacHub.new(dac.address);

        console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
        
        await dac.transfer(advertiser,100000*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        await dac.approve(dachub.address,20*dac,{from: advertiser});
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 20*dac);

        /*发布广告: id=12121.12122,12123 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
        var scanCount = 5;
        var mediaID1 = 12121;
        await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));

        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,50);

        /**追加扫码5次 */
        await dachub.updateAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));

        assert.equal((await dachub.mediaStrategy(mediaID1))[6],9);
    });

    it("Case7: 1个广告,限制5次,去使能分成,用户扫码无效,再使能,扫码有效", async function() {
        
        dac = await dacToken.new(100000*dac);
        dachub = await dacHub.new(dac.address);

        console.log('dacTotalSupply: '+(await dac.getdacTotalSupply()));
        
        await dac.transfer(advertiser,100000*dac);
        assert.equal((await dac.balanceOf(advertiser)).toNumber(), 100000*dac);

        await dac.approve(dachub.address,10*dac,{from: advertiser});
        assert.equal((await dac.allowance(advertiser,dachub.address)).toNumber(), 10*dac);

        /*发布广告: id=12121.12122,12123 扫码次数:5 每次给用户:1dac 其他分成比例: [50,30,20]*/ 
        var scanCount = 5;
        var mediaID1 = 12121;
        await dachub.publishAdvertise(advertiser,mediaID1,scanCount,1*dac,[50,30,20]);
        console.log('mediaStrategy: '+(await dachub.mediaStrategy(mediaID1)));

        await dachub.pauseIssue(mediaID1);
        
        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,100);
        assert.equal((await dac.balanceOf(user)).toNumber(), 0);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0);
        
        await dachub.enableIssue(mediaID1);
        
        /* 1次扫码 */
        await dachub.issueReward(user,screen,mediaID1,100);
        assert.equal((await dac.balanceOf(user)).toNumber(), 1*dac);
        assert.equal((await dac.balanceOf(paltform)).toNumber(), 0.5*dac);
        assert.equal((await dac.balanceOf(screen)).toNumber(), 0.3*dac);

    });
}); 
  